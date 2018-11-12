package game_server

import (
	"github.com/gorilla/websocket"
)

// соединение пользователя, заведомо валидное.
type UserConnection struct {
	Login      string
	Cookie     string
	Connection *websocket.Conn
}

// адрес игровой комнаты, уникальный для данного сервера.
type RoomId uint

// роль персонажа. при передаче состояния пользователю, если роли равны, персонаж
// называется синим, не равны - красным.
type RoleId uint8 // ∈ [1, 2]

// описание принадлежности к игре. Номер игровой комнаты и номер в игре,
// певый или второй игрок. Второй хранится на сервере в перевёрнутом состоянии.
type GameToСonnect struct {
	Room RoomId
	Role RoleId
}

// Список соединений, существующих в данный момент.
// используется для повторного подключения к той же игре, что и раньше.
// изменяется из конструктора/деструктора игровой комнаты.
type ProcessedPlayers map[string]GameToСonnect

// Оружие персонажа. Нападение на персонажа со флагом вызывает конец игры.
// Флаг не может нападать.
type Weapon string // ∈ ["stone", "scissors", "paper", "flag"]

// персонаж в представлении сервера.
type Сharacter struct {
	Role         RoleId
	Weapon       Weapon
	ShowedWeapon bool
}

// Карта в представлении сервера, координаты клеток 0 <= x <= 7, 0 <= y <= 6, для пустых клеток nil.
type Map [7][6]*Сharacter

//

// Необходимые глобальные объекты:
// ConnectionUpgrader
// Умеет принимать соединения,
// Проверять cookies,
// проверять подтягивать логин персонажа.

// LostAndFoundManager
// проверяет ProcessedPlayers на наличие комнаты для этого пользователя.
// возвращает соединение в комнату или замещает старое, или передаёт создателю комнат, если первый раз пришёл пользователь

// RoomCreator
// попарное вытаксивание соединений из очереди. Нумерация попарная ролей
// создание комнат.
// пометка соединений в ProcessedPlayers
// запуск надзирателя комнаты и слушающих события процессов

// 4 горутины на комнату, что изолируют соединение от игровой логики и подметы соединений менеджером потерь.

//    ╭─firstFrom─▶─╮      ╭─◀─SecondFrom─╮
// User1           GameMaster            User2
//    ╰─firstTo───◀─╯      ╰─▶─SecondTo───╯

// Мастер игры, содержит всю игровую логику, в один поток принимает / рассылает запросы, работает с картой.
// также содержит JSPN RPC сервер, вызывающий функции объекта комнаты.