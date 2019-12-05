package main

import (
	"fmt"
	"log"
	"net"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

func main() {
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

	fmt.Println("starting server at :8080")

	server.Serve(lis)
}
