package dispatcher

import (
	"context"
	"time"

	"github.com/aliok/websocket-channel/pkg/client/injection/client"
	websocketchannelinformer "github.com/aliok/websocket-channel/pkg/client/injection/informers/channels/v1alpha1/websocketchannel"
	websocketchannelreconciler "github.com/aliok/websocket-channel/pkg/client/injection/reconciler/channels/v1alpha1/websocketchannel"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing/pkg/channel"
	"knative.dev/eventing/pkg/channel/multichannelfanout"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	readTimeout   = 15 * time.Minute
	writeTimeout  = 15 * time.Minute
	port          = 8080
	finalizerName = "websocket-ch-dispatcher"
)

type envConfig struct {
	// TODO: change this environment variable to something like "PodGroupName".
	PodName       string `envconfig:"POD_NAME" required:"true"`
	ContainerName string `envconfig:"CONTAINER_NAME" required:"true"`
}

type NoopStatsReporter struct {
}

func (r *NoopStatsReporter) ReportEventCount(_ *channel.ReportArgs, _ int) error {
	return nil
}

func (r *NoopStatsReporter) ReportEventDispatchTime(_ *channel.ReportArgs, _ int, _ time.Duration) error {
	return nil
}

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(
	ctx context.Context,
	_ configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Panicw("Failed to process env var", zap.Error(err))
	}

	reporter := &NoopStatsReporter{}

	sh := multichannelfanout.NewMessageHandler(ctx, logger.Desugar(), channel.NewMessageDispatcher(logger.Desugar()), reporter)

	args := &webSocketMessageDispatcherArgs{
		Port:         port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      sh,
		Logger:       logger.Desugar(),
	}
	webSocketDispatcher := newMessageDispatcher(args)

	r := &Reconciler{
		multiChannelMessageHandler: sh,
		clientSet:                  client.Get(ctx).ChannelsV1alpha1(),
		reporter:                   reporter,
	}
	impl := websocketchannelreconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
		return controller.Options{SkipStatusUpdates: true, FinalizerName: finalizerName}
	})

	logging.FromContext(ctx).Info("Setting up event handlers")

	webSocketChannelInformer := websocketchannelinformer.Get(ctx)

	// Watch for channels.
	webSocketChannelInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool { return true },
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    impl.Enqueue,
				UpdateFunc: controller.PassNew(impl.Enqueue),
				DeleteFunc: r.deleteFunc,
			}})

	// Start the dispatcher.
	go func() {
		err := webSocketDispatcher.Start(ctx)
		if err != nil {
			logging.FromContext(ctx).Errorw("Failed stopping inMemoryDispatcher.", zap.Error(err))
		}
	}()

	return impl
}
