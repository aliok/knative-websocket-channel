package main

import (
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	"github.com/aliok/websocket-channel/pkg/reconciler/dispatcher"
)

func main() {
	ctx := signals.NewContext()

	sharedmain.MainWithContext(ctx, "websocket-channel-dispatcher",
		dispatcher.NewController,
	)
}
