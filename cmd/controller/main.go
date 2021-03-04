package main

import (
	"knative.dev/pkg/injection/sharedmain"

	controller "github.com/aliok/websocket-channel/pkg/reconciler/controller"
)

func main() {
	sharedmain.Main("websocketchannel-controller",
		controller.NewController,
	)
}
