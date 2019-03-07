package main

import (
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'
	"log"
	"net/http"
	"strconv"

	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/connection_upgrader"
	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/game_logic"
	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/websocket_test_page"
)

func main() {
	listenPort := flag.Uint16("listen-port", 8080, "listen port for websocket server")
	// authorisationServerPort := flag.Uint16("authorisation-port", 8081, "port for grpc connection to the authentication server")
	flag.Parse()
	// Инициализируем подсервер авторизации. connection_upgrader через него подтягивает login по cookie.

	// Инициализируем upgrader - он превращает соединения в websocket.
	upgrader := connectionUpgrader.NewConnectionUpgrader()
	roomsManager := game_logic.NewRoomsManager()
	go roomsManager.Run(upgrader.QueueToGame)
	http.HandleFunc("/game/v1/entrypoint", upgrader.HTTPEntryPoint)
	http.HandleFunc("/", websocket_test_page.WebSocketTestPage)
	portStr := strconv.Itoa(int(*listenPort))
	log.Println("Listening on :" + portStr)
	log.Print(http.ListenAndServe(":"+portStr, nil))
}
