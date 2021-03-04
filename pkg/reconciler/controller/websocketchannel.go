package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"

	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"

	pkgreconciler "knative.dev/pkg/reconciler"

	websocketchannelreconciler "github.com/aliok/websocket-channel/pkg/client/injection/reconciler/channels/v1alpha1/websocketchannel"
	listers "github.com/aliok/websocket-channel/pkg/client/listers/channels/v1alpha1"
)

type Reconciler struct {
	kubeClientSet kubernetes.Interface

	systemNamespace          string
	dispatcherImage          string
	websocketchannelLister   listers.WebSocketChannelLister
	websocketchannelInformer cache.SharedIndexInformer
	deploymentLister         appsv1listers.DeploymentLister
	serviceLister            corev1listers.ServiceLister
	endpointsLister          corev1listers.EndpointsLister
	serviceAccountLister     corev1listers.ServiceAccountLister
	roleBindingLister        rbacv1listers.RoleBindingLister
}

// Check that our Reconciler implements Interface
var _ websocketchannelreconciler.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, wsc *v1alpha1.WebSocketChannel) pkgreconciler.Event {

	// Make sure the dispatcher deployment exists and propagate the status to the Channel
	// For namespace-scope dispatcher, make sure configuration files exist and RBAC is properly configured.
	d, err := r.reconcileDispatcher(ctx, r.systemNamespace, wsc)
	if err != nil {
		logging.FromContext(ctx).Errorw("Failed to reconcile WebSocketChannel dispatcher", zap.Error(err))
		return err
	}
	wsc.Status.PropagateDispatcherStatus(&d.Status)



	// Make sure the dispatcher service exists and propagate the status to the Channel in case it does not exist.
	// We don't do anything with the service because it's status contains nothing useful, so just do
	// an existence check. Then below we check the endpoints targeting it.
	_, err = r.reconcileDispatcherService(ctx, r.systemNamespace, wsc)
	if err != nil {
		logging.FromContext(ctx).Errorw("Failed to reconcile WebSocketChannel dispatcher service", zap.Error(err))
		return err
	}
	wsc.Status.MarkServiceTrue()



	// Get the Dispatcher Service Endpoints and propagate the status to the Channel
	// endpoints has the same name as the service, so not a bug.
	e, err := r.endpointsLister.Endpoints(r.systemNamespace).Get(dispatcherName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Error("Endpoints do not exist for dispatcher service")
			wsc.Status.MarkEndpointsFailed("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist")
		} else {
			logging.FromContext(ctx).Error("Unable to get the dispatcher endpoints", zap.Error(err))
			wsc.Status.MarkEndpointsUnknown("DispatcherEndpointsGetFailed", "Failed to get dispatcher endpoints")
		}
		return err
	}
	if len(e.Subsets) == 0 {
		logging.FromContext(ctx).Error("No endpoints found for Dispatcher service", zap.Error(err))
		wsc.Status.MarkEndpointsFailed("DispatcherEndpointsNotReady", "There are no endpoints ready for Dispatcher service")
		return errors.New("there are no endpoints ready for Dispatcher service")
	}
	wsc.Status.MarkEndpointsTrue()




	// Reconcile the k8s service representing the actual Channel. It points to the Dispatcher service via
	// ExternalName
	svc, err := r.reconcileChannelService(ctx, r.systemNamespace, wsc)
	if err != nil {
		logging.FromContext(ctx).Errorw("Failed to reconcile channel service", zap.Error(err))
		return err
	}
	wsc.Status.MarkChannelServiceTrue()
	wsc.Status.SetAddress(apis.HTTP(network.GetServiceHostname(svc.Name, svc.Namespace)))



	// Ok, so now the Dispatcher Deployment & Service have been created, we're golden since the
	// dispatcher watches the Channel and where it needs to dispatch events to.
	logging.FromContext(ctx).Debugw("Reconciled WebSocketChannel", zap.Any("WebSocketChannel", wsc))
	return nil

}

func (r *Reconciler) reconcileDispatcher(ctx context.Context, dispatcherNamespace string, wsc *v1alpha1.WebSocketChannel) (*appsv1.Deployment, error) {
	d, err := r.deploymentLister.Deployments(dispatcherNamespace).Get(dispatcherName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			wsc.Status.MarkDispatcherFailed("DispatcherDeploymentDoesNotExist", "Dispatcher Deployment does not exist")
		} else {
			logging.FromContext(ctx).Error("Unable to get the dispatcher Deployment", zap.Error(err))
			wsc.Status.MarkDispatcherFailed("DispatcherDeploymentGetFailed", "Failed to get dispatcher Deployment")
		}
		return nil, newDeploymentWarn(err)
	}
	return d, nil
}

func (r *Reconciler) reconcileDispatcherService(ctx context.Context, dispatcherNamespace string, wsc *v1alpha1.WebSocketChannel) (*corev1.Service, error) {
	svc, err := r.serviceLister.Services(dispatcherNamespace).Get(dispatcherName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			wsc.Status.MarkServiceFailed("DispatcherServiceDoesNotExist", "Dispatcher Service does not exist")
		} else {
			logging.FromContext(ctx).Error("Unable to get the dispatcher service", zap.Error(err))
			wsc.Status.MarkServiceFailed("DispatcherServiceGetFailed", "Failed to get dispatcher service")
		}
		return nil, newServiceWarn(err)
	}
	return svc, nil
}

func (r *Reconciler) reconcileChannelService(ctx context.Context, dispatcherNamespace string, wsc *v1alpha1.WebSocketChannel) (*corev1.Service, error) {
	// Get the  Service and propagate the status to the Channel in case it does not exist.
	// We don't do anything with the service because it's status contains nothing useful, so just do
	// an existence check. Then below we check the endpoints targeting it.
	// We may change this name later, so we have to ensure we use proper addressable when resolving these.
	expected, err := newK8sService(wsc, externalService(dispatcherNamespace, dispatcherName))
	if err != nil {
		logging.FromContext(ctx).Error("failed to create the channel service object", zap.Error(err))
		wsc.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprint("Channel Service failed: ", err))
		return nil, err
	}

	channelSvcName := createChannelServiceName(wsc.Name)

	svc, err := r.serviceLister.Services(wsc.Namespace).Get(channelSvcName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			svc, err = r.kubeClientSet.CoreV1().Services(wsc.Namespace).Create(ctx, expected, metav1.CreateOptions{})
			if err != nil {
				logging.FromContext(ctx).Error("failed to create the channel service", zap.Error(err))
				wsc.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprint("Channel Service failed: ", err))
				return nil, err
			}
			return svc, nil
		}
		logging.FromContext(ctx).Error("Unable to get the channel service", zap.Error(err))
		wsc.Status.MarkChannelServiceUnknown("ChannelServiceGetFailed", fmt.Sprint("Unable to get the channel service: ", err))
		return nil, err
	} else if !equality.Semantic.DeepEqual(svc.Spec, expected.Spec) {
		svc = svc.DeepCopy()
		svc.Spec = expected.Spec

		svc, err = r.kubeClientSet.CoreV1().Services(wsc.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			logging.FromContext(ctx).Error("failed to update the channel service", zap.Error(err))
			wsc.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprint("Channel Service failed: ", err))
			return nil, err
		}
	}

	// Check to make sure that our channel owns this service and if not, complain.
	if !metav1.IsControlledBy(svc, wsc) {
		err := fmt.Errorf("websocketchannel: %s/%s does not own Service: %q", wsc.Namespace, wsc.Name, svc.Name)
		wsc.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprint("Channel Service failed: ", err))
		return nil, err
	}
	return svc, nil
}

func newDeploymentWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherDeploymentFailed", "Reconciling dispatcher Deployment failed with: %s", err)
}

func newServiceWarn(err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DispatcherServiceFailed", "Reconciling dispatcher Service failed: %s", err)
}
