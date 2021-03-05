package dispatcher

import (
	"context"
	"fmt"

	"github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	channelsv1 "github.com/aliok/websocket-channel/pkg/client/clientset/versioned/typed/channels/v1alpha1"
	reconcilerv1 "github.com/aliok/websocket-channel/pkg/client/injection/reconciler/channels/v1alpha1/websocketchannel"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/eventing/pkg/channel"
	"knative.dev/eventing/pkg/channel/fanout"
	"knative.dev/eventing/pkg/channel/multichannelfanout"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// Reconciler reconciles WebSocket Channels.
type Reconciler struct {
	multiChannelMessageHandler multichannelfanout.MultiChannelMessageHandler
	reporter                   channel.StatsReporter
	clientSet                  channelsv1.ChannelsV1alpha1Interface
}

// Check the interfaces Reconciler should implement
var (
	_ reconcilerv1.Interface         = (*Reconciler)(nil)
	_ reconcilerv1.ReadOnlyInterface = (*Reconciler)(nil)
)

func (r *Reconciler) ReconcileKind(ctx context.Context, wsc *v1alpha1.WebSocketChannel) reconciler.Event {
	if err := r.reconcile(ctx, wsc); err != nil {
		return err
	}

	// Then patch the subscribers to reflect that they are now ready to go
	return r.patchSubscriberStatus(ctx, wsc)
}

func (r *Reconciler) ObserveKind(ctx context.Context, wsc *v1alpha1.WebSocketChannel) reconciler.Event {
	return r.reconcile(ctx, wsc)
}

func (r *Reconciler) reconcile(ctx context.Context, wsc *v1alpha1.WebSocketChannel) reconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("WebSocketChannel", wsc))

	if !wsc.Status.IsReady() {
		logging.FromContext(ctx).Debug("WSC is not ready, skipping")
		return nil
	}

	config, err := newConfigForWebSocketChannel(wsc)
	if err != nil {
		logging.FromContext(ctx).Error("Error creating config for web socket channels", zap.Error(err))
		return err
	}

	// First grab the MultiChannelFanoutMessage handler
	handler := r.multiChannelMessageHandler.GetChannelHandler(config.HostName)
	if handler == nil {
		// No handler yet, create one.
		fanoutHandler, err := fanout.NewFanoutMessageHandler(
			logging.FromContext(ctx).Desugar(),
			channel.NewMessageDispatcher(logging.FromContext(ctx).Desugar()),
			config.FanoutConfig,
			r.reporter,
		)
		if err != nil {
			logging.FromContext(ctx).Error("Failed to create a new fanout.MessageHandler", err)
			return err
		}
		r.multiChannelMessageHandler.SetChannelHandler(config.HostName, fanoutHandler)
	} else {
		// Just update the config if necessary.
		haveSubs := handler.GetSubscriptions(ctx)

		// Ignore the closures, we stash the values that we can tell from if the values have actually changed.
		if diff := cmp.Diff(config.FanoutConfig.Subscriptions, haveSubs, cmpopts.IgnoreFields(kncloudevents.RetryConfig{}, "Backoff", "CheckRetry")); diff != "" {
			logging.FromContext(ctx).Info("Updating fanout config: ", zap.String("Diff", diff))
			handler.SetSubscriptions(ctx, config.FanoutConfig.Subscriptions)
		}
	}

	return nil
}

func (r *Reconciler) patchSubscriberStatus(ctx context.Context, wsc *v1alpha1.WebSocketChannel) error {
	after := wsc.DeepCopy()

	after.Status.Subscribers = make([]eventingduckv1.SubscriberStatus, 0)
	for _, sub := range wsc.Spec.Subscribers {
		after.Status.Subscribers = append(after.Status.Subscribers, eventingduckv1.SubscriberStatus{
			UID:                sub.UID,
			ObservedGeneration: sub.Generation,
			Ready:              corev1.ConditionTrue,
		})
	}
	jsonPatch, err := duck.CreatePatch(wsc, after)
	if err != nil {
		return fmt.Errorf("creating JSON patch: %w", err)
	}
	// If there is nothing to patch, we are good, just return.
	// Empty patch is [], hence we check for that.
	if len(jsonPatch) == 0 {
		return nil
	}

	patch, err := jsonPatch.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshaling JSON patch: %w", err)
	}
	patched, err := r.clientSet.WebSocketChannels(wsc.Namespace).Patch(ctx, wsc.Name, types.JSONPatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		return fmt.Errorf("Failed patching: %w", err)
	}
	logging.FromContext(ctx).Debugw("Patched resource", zap.Any("patch", patch), zap.Any("patched", patched))
	return nil
}

func newConfigForWebSocketChannel(wsc *v1alpha1.WebSocketChannel) (*multichannelfanout.ChannelConfig, error) {
	subs := make([]fanout.Subscription, len(wsc.Spec.Subscribers))

	for i, sub := range wsc.Spec.Subscribers {
		conf, err := fanout.SubscriberSpecToFanoutConfig(sub)
		if err != nil {
			return nil, err
		}
		subs[i] = *conf
	}

	return &multichannelfanout.ChannelConfig{
		Namespace: wsc.Namespace,
		Name:      wsc.Name,
		HostName:  wsc.Status.Address.URL.Host,
		FanoutConfig: fanout.Config{
			AsyncHandler:  true,
			Subscriptions: subs,
		},
	}, nil
}
