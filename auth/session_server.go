package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	grpcPort = flag.Int("grpc", 8080, "listen addr")
)

func main() {
	flag.Parse()

	port := strconv.Itoa(*grpcPort)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Println(err)
	}

	server := grpc.NewServer()

	session.RegisterAuthCheckerServer(server, session.NewSessionManager(redisConn))

	// sessionManager.Create(context.Background(), &session.Session{
	// 	Username: "login", Avatar: "default",
	// })

	config := consulapi.DefaultConfig()
	config.Address = utils.ConsulAddr
	consul, err := consulapi.NewClient(config)

	serviceID := "SAPI_127.0.0.1:" + port

	err = consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    "session-api",
		Port:    *grpcPort,
		Address: "127.0.0.1",
	})
	if err != nil {
		fmt.Println("cant add service to consul", err)
		return
	}
	fmt.Println("registered in consul", serviceID)

	defer func() {
		err := consul.Agent().ServiceDeregister(serviceID)
		if err != nil {
			fmt.Println("cant add service to consul", err)
			return
		}
		fmt.Println("deregistered in consul", serviceID)
	}()

	fmt.Println("starting server at :8080")

	go server.Serve(lis)
}
