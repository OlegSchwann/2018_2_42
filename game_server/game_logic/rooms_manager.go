package game_logic

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/user_connection"
)

// адрес игровой комнаты, уникальный для данного сервера.
type RoomId uint

func (ri RoomId) String() string {
	return strconv.FormatUint(uint64(ri), 10)
}

// роль персонажа. при передаче состояния пользователю, если роли равны, персонаж
// называется синим, не равны - красным.
type RoleId uint8 // ∈ [0, 1]

func (ri RoleId) String() string {
	if ri == 0 {
		return "0"
	}
	return "1"
}

// описание принадлежности к игре. Номер игровой комнаты и номер в игре,
// певый или второй игрок. Второй хранится на сервере в перевёрнутом состоянии.
type GameToConnect struct {
	Room RoomId
	Role RoleId
}

// Оружие персонажа. Нападение на персонажа со флагом вызывает конец игры. Флаг не может нападать.
type Weapon string // ∈ ["stone", "scissors", "paper", "flag"]

func NewWeapon(key string) (weapon Weapon, err error) {
	switch key {
	case "rock":
		fallthrough
	case "scissors":
		fallthrough
	case "paper":
		fallthrough
	case "flag":
		weapon = Weapon(key)
	default:
		err = errors.New("'" + key + "' ∉ ['rock', 'scissors', 'paper', 'flag']")
	}
	return
}

// true если превосходит передаваемое значение, false
func (w *Weapon) IsExceed(rival Weapon) (exceed bool) {
	switch *w {
	case "rock":
		exceed = rival == "scissors"
	case "scissors":
		exceed = rival == "paper"
	case "paper":
		exceed = rival == "rock"
	}
	return
}

// Персонаж в представлении сервера.
type Сharacter struct {
	Role         RoleId
	Weapon       Weapon
	ShowedWeapon bool
}

func (c *Сharacter) String() (str string) {
	if c == nil {
		str = "            "
	} else {
		if c.Role == 0 {
			str += "0 "
		} else {
			str += "1 "
		}
		switch c.Weapon {
		case "rock":
			str += "rock     "
		case "scissors":
			str += "scissors "
		case "paper":
			str += "paper    "
		case "flag":
			str += "flag     "
		}
		if c.ShowedWeapon {
			str += "+"
		} else {
			str += "-"
		}
	}
	return
}

// Карта в представлении сервера, координаты клеток 0 <= x <= 41, для пустых клеток nil.
//[ 0,  1,  2,  3,  4,  5,  6,
//  7,  8,  9, 10, 11, 12, 13,
// 14, 15, 16, 17, 18, 19, 20,
// 21, 22, 23, 24, 25, 26, 27,
// 28, 29, 30, 31, 32, 33, 34,
// 35, 36, 37, 38, 39, 40, 41]
type Map [42]*Сharacter

func (m Map) String() (str string) { // implement fmt.Stringer interface, called fmt.Print()
	separator := "├────────────┼────────────┼────────────┼────────────┼────────────┼────────────┼────────────┤\n"
	row := func(i int) string {
		return fmt.Sprint("│", m[i], "│", m[i+1], "│", m[i+2], "│", m[i+3], "│", m[i+4], "│", m[i+5], "│", m[i+6], "│\n")
	}
	str = "┌────────────┬────────────┬────────────┬────────────┬────────────┬────────────┬────────────┐\n" +
		row(0) + separator + row(7) + separator + row(14) + separator + row(21) + separator + row(28) + separator + row(35) +
		"└────────────┴────────────┴────────────┴────────────┴────────────┴────────────┴────────────┘\n"
	return
}

// Паттерн актор: горутина, распоряжающаяся этим классом запущена из main,
// живёт всё время работы в единственном экземпляре. Блокирующе читает из
// канала connection_upgrader.ConnectionUpgrader.QueueToGame, берёт пользователей
// по одному, проверяет ProcessedPlayers на наличие комнаты для этого пользователя.
// возвращает соединение в комнату или замещает старое, или, если игрок пришёл первый раз,
// добавляет его в создаваемую комнату, помечая соединения в ProcessedPlayers.
type RoomsManager struct {
	// Список соединений, существующих в данный момент.
	// используется для повторного подключения к той же игре, что и раньше.
	// изменяется из конструктора/деструктора игровой комнаты.
	// Ключ - login пользователя.
	ProcessedPlayers map[string]GameToConnect
	// Игровые комнаты.
	Rooms map[RoomId]*Room
	// последний номер созданной комнаты, что бы поддерживать уникальность номеров
	RoomNumber RoomId
	// пользователь, ждущий подключения другого пользователя.
	WaitingConnection *user_connection.UserConnection
	// канал "требование удаления"
	// комната передаёт сюда собственный RoomId, и комната, оба соединения удаляется из RoomManager.
	CompletedRooms chan RoomId
}

func NewRoomsManager() (roomsManager *RoomsManager) {
	roomsManager = &RoomsManager{
		ProcessedPlayers: make(map[string]GameToConnect),
		Rooms:            make(map[RoomId]*Room),
		CompletedRooms:   make(chan RoomId, 5),
	}
	return
}

func (rm *RoomsManager) Run(connectionQueue chan *user_connection.UserConnection) {
	for connectionQueue != nil && rm.CompletedRooms != nil {
		select { // https://stackoverflow.com/questions/13666253/breaking-out-of-a-select-statement-when-all-channels-are-closed
		case RoomId, ok := <-rm.CompletedRooms:
			if ok {
				rm.processRoomRemoval(RoomId)
			} else {
				rm.CompletedRooms = nil
			}
		case connection, ok := <-connectionQueue:
			if ok {
				rm.processUserAddition(connection)
			} else {
				connectionQueue = nil
			}
		}
	}
	return
}

func (rm *RoomsManager) processUserAddition(connection *user_connection.UserConnection) {
	// если пользователь с таким cookie sessionid уже играет
	game, ok := rm.ProcessedPlayers[connection.Token]
	if ok {
		// то восстанавливаем соединение.
		log.Printf("Reconnect user = '%s' in role %d to room %d", connection.Token, game.Role, game.Room)
		rm.Rooms[game.Room].Reconnect(connection, game.Role)
		return
	}

	if rm.WaitingConnection == nil {
		log.Printf("Set connection user = '%s' as waiting", connection.Token)
		rm.WaitingConnection = connection
		return
	}

	// добавление в новую комнату 2-х соединений и регистрация пользователей,
	// как находящихся в процессе игры.
	log.Printf("create room %d user0 = '%s', user1 = '%s'", rm.RoomNumber, rm.WaitingConnection.Token, connection.Token)

	rm.Rooms[rm.RoomNumber] = NewRoom(rm.WaitingConnection, connection, rm.CompletedRooms, rm.RoomNumber)
	rm.ProcessedPlayers[rm.WaitingConnection.Token] = GameToConnect{
		Room: rm.RoomNumber,
		Role: 0,
	}
	rm.ProcessedPlayers[connection.Token] = GameToConnect{
		Room: rm.RoomNumber,
		Role: 1,
	}
	rm.WaitingConnection = nil
	rm.RoomNumber++
	return
}

func (rm *RoomsManager) processRoomRemoval(roomId RoomId) (err error) {
	room, ok := rm.Rooms[roomId]
	if !ok {
		err = errors.New("attempt to delete non-existing room by RoomId = " + roomId.String())
		log.Print(err)
		return
	}
	delete(rm.ProcessedPlayers, room.User0.Token)
	delete(rm.ProcessedPlayers, room.User1.Token)
	delete(rm.Rooms, roomId)
	log.Print("room RoomId = " + roomId.String() + " successfully removed")
	return
}
