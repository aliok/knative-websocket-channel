package dispatcher

import (
	"context"
	"time"

	"go.uber.org/zap"
	"knative.dev/eventing/pkg/channel/multichannelfanout"
	"knative.dev/eventing/pkg/kncloudevents"
)

type webSocketMessageDispatcher struct {
	handler              multichannelfanout.MultiChannelMessageHandler
	httpBindingsReceiver *kncloudevents.HTTPMessageReceiver
	writeTimeout         time.Duration
	logger               *zap.Logger
}

type webSocketMessageDispatcherArgs struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Handler      multichannelfanout.MultiChannelMessageHandler
	Logger       *zap.Logger
}

func newMessageDispatcher(args *webSocketMessageDispatcherArgs) *webSocketMessageDispatcher {
	// TODO set read timeouts?
	bindingsReceiver := kncloudevents.NewHTTPMessageReceiver(args.Port)

	dispatcher := &webSocketMessageDispatcher{
		handler:              args.Handler,
		httpBindingsReceiver: bindingsReceiver,
		logger:               args.Logger,
		writeTimeout:         args.WriteTimeout,
	}

	return dispatcher
}

func (d *webSocketMessageDispatcher) Start(ctx context.Context) error {
	return d.httpBindingsReceiver.StartListen(kncloudevents.WithShutdownTimeout(ctx, d.writeTimeout), d.handler)
}
