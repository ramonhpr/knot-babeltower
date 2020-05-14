package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/CESARBR/knot-babeltower/internal/config"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/server"
	thingControllers "github.com/CESARBR/knot-babeltower/pkg/thing/controllers"
	thingDeliveryAMQP "github.com/CESARBR/knot-babeltower/pkg/thing/delivery/amqp"
	thingDeliveryHTTP "github.com/CESARBR/knot-babeltower/pkg/thing/delivery/http"
	thingInteractors "github.com/CESARBR/knot-babeltower/pkg/thing/interactors"
	userControllers "github.com/CESARBR/knot-babeltower/pkg/user/controllers"
	userDeliveryHTTP "github.com/CESARBR/knot-babeltower/pkg/user/delivery/http"
	userInteractors "github.com/CESARBR/knot-babeltower/pkg/user/interactors"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
)

func monitorSignals(sigs chan os.Signal, quit chan bool, logger logging.Logger) {
	signal := <-sigs
	logger.Infof("signal %s received", signal)
	quit <- true
}

// Main will be used for unit tests
func Main(config config.Config, quit chan bool, startedChan chan bool) {
	logrus := logging.NewLogrus(config.Logger.Level)

	logger := logrus.Get("Main")
	logger.Info("starting KNoT Babeltower")

	// Signal Handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go monitorSignals(sigs, quit, logger)

	// AMQP
	amqpStartedChan := make(chan bool, 1)
	amqp := network.NewAmqp(config.RabbitMQ.URL, logrus.Get("Amqp"))

	// AMQP Publishers
	clientPublisher := thingDeliveryAMQP.NewMsgClientPublisher(logrus.Get("ClientPublisher"), amqp.GetSender())
	commandSender := thingDeliveryAMQP.NewCommandSender(logrus.Get("Command Sender"), amqp.GetSender())

	// Services
	userProxy := userDeliveryHTTP.NewUserProxy(logrus.Get("UserProxy"), config.Users.Hostname, config.Users.Port)
	thingProxy := thingDeliveryHTTP.NewThingProxy(logrus.Get("ThingProxy"), config.Things.Hostname, config.Things.Port)

	// Interactors
	createUser := userInteractors.NewCreateUser(logrus.Get("CreateUser"), userProxy)
	createToken := userInteractors.NewCreateToken(logrus.Get("CreateToken"), userProxy)
	thingInteractor := thingInteractors.NewThingInteractor(logrus.Get("ThingInteractor"), clientPublisher, thingProxy)

	// Controllers
	thingController := thingControllers.NewThingController(logrus.Get("ThingController"), thingInteractor, commandSender)
	userController := userControllers.NewUserController(logrus.Get("UserController"), createUser, createToken)

	// Server
	serverStartedChan := make(chan bool, 1)
	http := server.NewServer(config.Server.Port, logrus.Get("Server"), userController)

	// AMQP Handler
	msgStartedChan := make(chan bool, 1)
	msgHandler := server.NewMsgHandler(logrus.Get("MsgHandler"), amqp.GetReceiver(), thingController)

	// Start goroutines
	go amqp.Start(amqpStartedChan)
	go http.Start(serverStartedChan)

	// Main loop
	for {
		select {
		case started := <-serverStartedChan:
			if started {
				logger.Info("server started")
			}
		case started := <-amqpStartedChan:
			if started {
				logger.Info("AMQP connection started")
				go msgHandler.Start(msgStartedChan)
			}
		case started := <-msgStartedChan:
			if started {
				logger.Info("message handler started")
				startedChan <- true
			} else {
				quit <- true
			}
		case <-quit:
			msgHandler.Stop()
			amqp.Stop()
			http.Stop()
			os.Exit(0)
		}
	}
}

func main() {
	Main(config.Load(), make(chan bool, 1), make(chan bool, 1))
}
