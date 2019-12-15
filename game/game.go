package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/handlers"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	repository "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/repositoryLeaders"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	consulapi "github.com/hashicorp/consul/api"

	_ "github.com/go-sql-driver/mysql"
)

var (
	consulAddr = flag.String("addr", "127.0.0.1:8500", "consul addr (8500 in original consul)")
)

var (
	consul       *consulapi.Client
	nameResolver *testNameResolver
	servers      []string
)

const (
	ListenAddr = "172.26.112.3:443"
	FrontIP    = "http://93.171.139.195:780"
	FrontIPNew = "https://www.pokertodo.ru:743"
)

func main() {
	flag.Parse()

	var err error
	config := consulapi.DefaultConfig()
	config.Address = *consulAddr
	consul, err = consulapi.NewClient(config)

	health, _, err := consul.Health().Service("session-api", "", false, nil)
	if err != nil {
		log.Fatalf("cant get alive services")
	}

	servers = []string{}
	for _, item := range health {
		addr := item.Service.Address +
			":" + strconv.Itoa(item.Service.Port)
		servers = append(servers, addr)
	}

	nameResolver = &testNameResolver{
		addr: servers[0],
	}
	log.Println(nameResolver)

	grcpConn, err := grpc.Dial(
		servers[0],
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBalancer(grpc.RoundRobin(nameResolver)),
	)
	if err != nil {
		log.Fatalf("cant connect to grpc")
	}
	defer grcpConn.Close()

	if len(servers) > 1 {
		var updates []*naming.Update
		for i := 1; i < len(servers); i++ {
			updates = append(updates, &naming.Update{
				Op:   naming.Add,
				Addr: servers[i],
			})
		}
		nameResolver.w.inject(updates)
	}

	handlers.SessManager = session.NewAuthCheckerClient(grcpConn)

	go runOnlineServiceDiscovery(servers)

	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method}] ${remote_ip}, ${uri} ${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{FrontIP, FrontIPNew},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	leaderBoardDB := repository.NewUserMemoryRepository()

	handlers := &handlers.HandlersGame{
		Usecase: leaderBoardDB,
	}

	e.GET("/rooms/", handlers.GetRooms)
	e.GET("/multiplayer/", handlers.WsHandler)
	e.GET("/leaderboard/", handlers.LeaderBoardTopHandler)

	e.Logger.Fatal(e.StartTLS(ListenAddr, "cert.crt", "key.crt"))
}

func runOnlineServiceDiscovery(servers []string) {
	currAddrs := make(map[string]struct{}, len(servers))
	for _, addr := range servers {
		currAddrs[addr] = struct{}{}
	}
	ticker := time.Tick(5 * time.Second)
	for _ = range ticker {
		health, _, err := consul.Health().Service("session-api", "", false, nil)
		if err != nil {
			log.Fatalf("cant get alive services")
		}

		newAddrs := make(map[string]struct{}, len(health))
		for _, item := range health {
			addr := item.Service.Address +
				":" + strconv.Itoa(item.Service.Port)
			newAddrs[addr] = struct{}{}
		}

		var updates []*naming.Update
		// проверяем что удалилось
		for addr := range currAddrs {
			if _, exist := newAddrs[addr]; !exist {
				updates = append(updates, &naming.Update{
					Op:   naming.Delete,
					Addr: addr,
				})
				delete(currAddrs, addr)
				fmt.Println("remove", addr)
			}
		}
		// проверяем что добавилось
		for addr := range newAddrs {
			if _, exist := currAddrs[addr]; !exist {
				updates = append(updates, &naming.Update{
					Op:   naming.Add,
					Addr: addr,
				})
				currAddrs[addr] = struct{}{}
				fmt.Println("add", addr)
			}
		}
		if len(updates) > 0 {
			nameResolver.w.inject(updates)
		}
	}
}
