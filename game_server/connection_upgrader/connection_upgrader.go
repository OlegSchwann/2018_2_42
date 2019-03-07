package connectionUpgrader

import (
	"github.com/bxcodec/faker"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"

	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/types"
	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/user_connection"
)

// ConnectionUpgrader ответственен за превращение http соединения в игрока - пользователя.
type ConnectionUpgrader struct {
	// Настройки WebSocket.
	upgrader websocket.Upgrader
	// Канал, в который помещаются соединения с пользователем, что бы передать их в RoomManager
	QueueToGame chan *user_connection.UserConnection
}

// Фабричная функция ConnectionUpgrader.
func NewConnectionUpgrader() (cu *ConnectionUpgrader) {
	cu = &ConnectionUpgrader{
		upgrader: websocket.Upgrader{
			HandshakeTimeout: time.Duration(1 * time.Second),
			CheckOrigin: func(r *http.Request) bool { // Токен не проверяется.
				return true
			},
			EnableCompression: true,
		},
		QueueToGame: make(chan *user_connection.UserConnection, 50),
	}
	return
}

func (cu *ConnectionUpgrader) getAnonUserInfo() (login string, avatar string, err error) {
	username := struct {
		UserName string `faker:"username"`
	}{}

	err = faker.FakeData(&username)
	// создаёт случайный логин
	login = username.UserName
	avatar = "/images/default.png"
	return
}

// HTTPEntryPoint - входная точка для http соединения.
// Запускается в разных горутинах, только читает из класса.
// Проводит upgrade соединения и проверку cookie полззователя.
func (cu *ConnectionUpgrader) HTTPEntryPoint(w http.ResponseWriter, r *http.Request) {
	log.Printf("New connection: %#v", r)
	// Проверяет SessionId из cookie.
	sessionID, err := r.Cookie("SessionId")
	if err != nil {
		response, _ := types.ServerResponse{
			Status:  "forbidden",
			Message: "missing_sessionid_cookie",
		}.MarshalJSON()
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(response)
		_ = r.Body.Close()
		return
	}

	// Меняет протокол.
	WSConnection, err := cu.upgrader.Upgrade(w, r, nil)
	if err != nil {
		response, _ := types.ServerResponse{
			Status:  "bad request",
			Message: "error on upgrade connection: " + err.Error(),
		}.MarshalJSON()
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(response)
		_ = r.Body.Close()
		return
	}

	login, avatar, _ := cu.getAnonUserInfo()

	connection := &user_connection.UserConnection{
		Login:      login,
		Avatar:     avatar,
		Token:      sessionID.Value,
		Connection: WSConnection,
	}
	cu.QueueToGame <- connection
	return
}
