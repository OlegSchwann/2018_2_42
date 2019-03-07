package game_logic

import (
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"

	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/types"
)

func (r *Room) GameMaster() {
	log.Printf("start GameMaster for room: %#v", *r)
	var message []byte
	var role RoleId
gameLoop:
	for {
		select {
		case <-r.TimeoutTimer.C:
			r.Stop()
			r.Remove()
			log.Printf("room = " + r.OwnNumber.String() + ", time out!")
			break gameLoop
		case message = <-r.Messaging.User0From:
			role = 0
			log.Printf("message came from the User0: " + string(message))
		case message = <-r.Messaging.User1From:
			role = 1
			log.Printf("message came from the User1: " + string(message))
		}
		r.TimeoutTimer.Reset(timeForMove)

		event := types.Event{}
		err := event.UnmarshalJSON(message)
		if err != nil {
			response, _ := types.ErrorMessage("error while parsing first level: " + err.Error()).MarshalJSON()
			response, _ = types.Event{
				Method:    "error_message",
				Parameter: response,
			}.MarshalJSON()
			if role == 0 {
				r.Messaging.User0To <- response
			} else {
				r.Messaging.User1To <- response
			}
			continue
		}
		if event.Method == "upload_map" {
			err := r.UploadMap(role, event.Parameter)
			if err != nil {
				response, _ := types.ErrorMessage("error while process 'upload_map': " + err.Error()).MarshalJSON()
				response, _ = types.Event{
					Method:    "error_message",
					Parameter: response,
				}.MarshalJSON()
				if role == 0 {
					r.Messaging.User0To <- response
				} else {
					r.Messaging.User1To <- response
				}
				if r.User0UploadedCharacters && r.User1UploadedCharacters {
					r.DownloadMap(role)
				}
			}
			continue
		}
		if event.Method == "attempt_go_to_cell" {
			gameover, err := r.AttemptGoToCell(role, event.Parameter)
			if err != nil {
				response, _ := types.ErrorMessage("error while process 'attempt_go_to_cell': " + err.Error()).MarshalJSON()
				response, _ = types.Event{
					Method:    "error_message",
					Parameter: response,
				}.MarshalJSON()
				if role == 0 {
					r.Messaging.User0To <- response
				} else {
					r.Messaging.User1To <- response
				}
				if r.User0UploadedCharacters && r.User1UploadedCharacters {
					r.DownloadMap(role)
				}
			}
			if gameover {
				// к этому моменту эже все данные должны быть отправлены. только сетевые вопросы и остановка всех 5-и горутин.
				r.Stop()
				// отрегистирует в Rooms.
				r.Remove()
				break gameLoop
			}
			continue
		}
		if event.Method == "reassign_weapons" {
			err = r.ReassignWeapons(role, event.Parameter)
			if err != nil {
				response, _ := types.ErrorMessage("error while process 'reassign_weapons': " + err.Error()).MarshalJSON()
				response, _ = types.Event{
					Method:    "error_message",
					Parameter: response,
				}.MarshalJSON()
				if role == 0 {
					r.Messaging.User0To <- response
				} else {
					r.Messaging.User1To <- response
				}
				if r.User0UploadedCharacters && r.User1UploadedCharacters {
					r.DownloadMap(role)
				}
			}
			continue
		}
		// если ни один из трёх методов не отработал, прислали неверный метод, кидаем ошибку
		response, _ := types.Event{
			Method: "error_message",
			Parameter: easyjson.RawMessage("unknown method '" + event.Method + "', " +
				"available only ['attempt_go_to_cell', 'upload_map', 'reassign_weapons']."),
		}.MarshalJSON()
		if role == 0 {
			r.Messaging.User0To <- response
		} else {
			r.Messaging.User1To <- response
		}
	}
	log.Print("GameMaster for room = " + r.OwnNumber.String() + " correctly completed.")
	return
}

// ответственность: загружает данные от пользователя, начинает игру
func (r *Room) UploadMap(role RoleId, message easyjson.RawMessage) (err error) {
	var uploadedMap types.UploadMap
	err = uploadedMap.UnmarshalJSON(message)
	if err != nil {
		err = errors.Wrap(err, "in json.Unmarshal message into types.UploadMap: ")
		return
	}

	err = r.uploadMap(role, uploadedMap)
	if err != nil {
		return
	}

	if r.User0UploadedCharacters && r.User1UploadedCharacters {
		// Отсылает карту
		r.DownloadMap(0)
		r.DownloadMap(1)
		// Отсылает логин соперника
		r.YourRival(0)
		r.YourRival(1)
		// Отправляет чей ход
		r.YourTurn(0)
		r.YourTurn(1)
	}
	return
}

func (r *Room) uploadMap(role RoleId, uploadedMap types.UploadMap) (err error) {
	if role == 0 {
		if !r.User0UploadedCharacters {
			// uploadedMap.Weapons для клеток 13 12 11 10 9 8 7 6 5 4 3 2 1 0
			var numberOfFlags int
			for i := 0; i < 14; i++ {
				j := 13 - i
				var weapon Weapon
				weapon, err = NewWeapon(uploadedMap.Weapons[i])
				if err != nil {
					err = errors.Wrap(err, "in NewWeapon: ")
					return
				}
				if weapon == "flag" {
					numberOfFlags++
				}
				r.Map[j] = &Сharacter{
					Role:   role,
					Weapon: weapon,
				}
			}
			if numberOfFlags != 1 {
				err = errors.New("map must contain exactly one flag, but " +
					strconv.Itoa(numberOfFlags) + " found")
				return
			}
			r.User0UploadedCharacters = true
		} else {
			err = errors.New("characters already loaded")
			return
		}
	} else {
		if !r.User1UploadedCharacters {
			// 28 29 30 31 32 33 34 35 36 37 38 39 40 41
			var numberOfFlags int
			for i := 0; i < 14; i++ {
				j := 28 + i
				var weapon Weapon
				weapon, err = NewWeapon(uploadedMap.Weapons[i])
				if err != nil {
					err = errors.Wrap(err, "in NewWeapon: ")
					return
				}
				if weapon == "flag" {
					numberOfFlags++
				}
				r.Map[j] = &Сharacter{
					Role:   role,
					Weapon: weapon,
				}
			}
			if numberOfFlags != 1 {
				err = errors.New("map must contain exactly one flag, but " +
					strconv.Itoa(int(numberOfFlags)) + " found")
				return
			}
			r.User1UploadedCharacters = true
		} else {
			err = errors.New("characters already loaded")
			return
		}
	}
	return
}

// ответственность: отправляет карту на клиент, не изменяет карту.
func (r *Room) DownloadMap(role RoleId) {
	downloadMap := types.DownloadMap{}
	for i := 0; i < len(r.Map); i++ {
		if r.Map[i] == nil {
			continue
		}
		var cell = &types.MapCell{}
		// true, если собственный персонаж
		cell.User = r.Map[i].Role == role
		// оружие видно только если это собственный игрок или противник показал оружие.
		if r.Map[i].Role == role || r.Map[i].ShowedWeapon {
			weapon := string(r.Map[i].Weapon)
			cell.Weapon = &weapon
		}
		downloadMap[i] = cell
	}
	if role == 0 {
		downloadMap.Rotate()
	}
	parameter, _ := downloadMap.MarshalJSON()
	response, _ := types.Event{
		Method:    "download_map",
		Parameter: parameter,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: отправляет описание соперника, не изменяет карту.
func (r *Room) YourRival(role RoleId) {
	if role == 0 {
		response, _ := types.YourRival(r.User1.Login).MarshalJSON()
		response, _ = types.Event{
			Method:    "your_rival",
			Parameter: []byte(response),
		}.MarshalJSON()
		r.Messaging.User1To <- response
	} else {
		response, _ := types.YourRival(r.User0.Login).MarshalJSON()
		response, _ = types.Event{
			Method:    "your_rival",
			Parameter: []byte(response),
		}.MarshalJSON()
		r.Messaging.User0To <- response
	}
	return
}

// ответственность: отправляет стат чей ход, не изменяет карту.
func (r *Room) YourTurn(role RoleId) {
	var response []byte
	if types.YourTurn(r.UserTurnNumber == role) {
		response = []byte("true")
	} else {
		response = []byte("false")
	}
	response, _ = types.Event{
		Method:    "your_turn",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: принимает данные от пользователя, обрабатывает с учётом состояния,
// изменяет согласно игровой механике карту (фактически содержит всю игру в себе 😮)
// вызывает функции, отправляющие запросы.
func (r *Room) AttemptGoToCell(role RoleId, message easyjson.RawMessage) (gameOver bool, err error) {
	var attemptGoToCell types.AttemptGoToCell
	err = attemptGoToCell.UnmarshalJSON(message)
	if err != nil {
		err = errors.Wrap(err, "in json.Unmarshal message into types.attemptGoToCell: ")
		return
	}
	err = attemptGoToCell.Check()
	if err != nil {
		err = errors.Wrap(err, "invalid coordinates: ")
		return
	}
	if role == 0 {
		attemptGoToCell.Rotate()
	}
	gameOver, err = r.AttemptGoToCellLogic(role, attemptGoToCell.From, attemptGoToCell.To)
	return
}

func (r *Room) AttemptGoToCellLogic(role RoleId, from int, to int) (gameOver bool, err error) {
	// Что бы пользователю можно было сделать ход, нужно,
	// что бы персонажи были загружены обоими игроками,
	// не было спора про перевыбор оружия в данный момент неоконченного
	// и был ход этого игрока.
	if r.UserTurnNumber != role {
		err = errors.New("it's not your turn now")
		return
	}
	if !r.User0UploadedCharacters || !r.User1UploadedCharacters {
		err = errors.New("The map is not loaded yet. Wait for it.")
		return
	}
	if r.WeaponReElection.WaitingForIt {
		err = errors.New("At that moment you still need to reassign the weapon.")
		return
	}
	if r.Map[from] == nil {
		err = errors.New("there is no character at " + strconv.Itoa(from))
		return
	}
	if r.Map[from].Role != role {
		err = errors.New("this is not your character at " + strconv.Itoa(from))
		return
	}
	// Тут точно существующий персонаж, принадлежащий игроку.
	// Сервер смотрит, куда перемещается персонаж и, если целевая клетка пуста,
	// перемещает персонажа на сервере и клиентах.
	if r.Map[to] == nil {
		r.Map[to], r.Map[from] = r.Map[from], nil
		if r.UserTurnNumber == 0 {
			r.UserTurnNumber = 1
		} else {
			r.UserTurnNumber = 0
		}
		r.MoveCharacter(0, from, to)
		r.MoveCharacter(1, from, to)
		r.YourTurn(0)
		r.YourTurn(1)
		return
	}
	// если в целевой клетке ты
	if r.Map[to].Role == role {
		err = errors.New("attempt to attack yourself")
		return
	}
	// проверяем, нет ли там флага
	if r.Map[to].Weapon == "flag" {
		log.Print("game over in room = " + r.OwnNumber.String())
		r.Gameover(0, role, from, to)
		r.Gameover(1, role, from, to)
		gameOver = true
		return
	}
	// проверяем победу над обычным оружием.
	if r.Map[from].Weapon.IsExceed(r.Map[to].Weapon) {
		winnerWeapon := r.Map[from].Weapon
		loserWeapon := r.Map[to].Weapon
		// двигаем персонажа
		r.Map[to] = r.Map[from]
		r.Map[from] = nil
		// ставим, что оружие победителя спалилось.
		r.Map[to].ShowedWeapon = true
		// меняем ход
		if r.UserTurnNumber == 0 {
			r.UserTurnNumber = 1
		} else {
			r.UserTurnNumber = 0
		}
		// отсылаем изменения.
		r.Attack(0, from, winnerWeapon, to, loserWeapon)
		r.Attack(1, from, winnerWeapon, to, loserWeapon)
		// отсылаем смену хода
		r.YourTurn(0)
		r.YourTurn(1)
		return
	}
	// проверяем поражение
	if r.Map[to].Weapon.IsExceed(r.Map[from].Weapon) {
		winnerWeapon := r.Map[to].Weapon
		loserWeapon := r.Map[from].Weapon
		// убираем проигравшего нападавшего персонажа, победитель передвигается на клетку проигравшего.
		r.Map[from] = nil
		// ставим, что оружие победителя спалилось.
		r.Map[to].ShowedWeapon = true
		// меняем ход
		if r.UserTurnNumber == 0 {
			r.UserTurnNumber = 1
		} else {
			r.UserTurnNumber = 0
		}
		// отсылаем изменения.
		r.Attack(0, to, winnerWeapon, from, loserWeapon)
		r.Attack(1, to, winnerWeapon, from, loserWeapon)
		// отсылаем смену хода
		r.YourTurn(0)
		r.YourTurn(1)
		return
	}
	// проверяем, что одинаковое оружие
	if r.Map[to].Weapon == r.Map[from].Weapon {
		fmt.Print(r.Map)

		// запускаем процедуру перевыбора.
		r.WeaponReElection.WaitingForIt = true
		r.WeaponReElection.User0ReElect = false
		r.WeaponReElection.User1ReElect = false
		r.WeaponReElection.AttackingCharacter = from
		r.WeaponReElection.AttackedCharacter = to

		// просим игроков перевыбрать оружие для своего персонажа, ход не меняется.
		if r.UserTurnNumber == 0 {
			r.WeaponChangeRequest(0, from)
			r.WeaponChangeRequest(1, to)
		} else {
			r.WeaponChangeRequest(1, from)
			r.WeaponChangeRequest(0, to)
		}
		return
	}
	return
}

// ответственность: проводит загружает перевыбранное оружие,
// вызывает AttemptGoToCell снова, как бужто перевыбора небыло.
func (r *Room) ReassignWeapons(role RoleId, message easyjson.RawMessage) (err error) {
	reassignWeapons := types.ReassignWeapons{}
	err = reassignWeapons.UnmarshalJSON(message)
	if err != nil {
		err = errors.Wrap(err, "parsing error: ")
		return
	}
	weapon, err := NewWeapon(reassignWeapons.NewWeapon)
	if err != nil {
		err = errors.Wrap(err, "incorrect weapon: ")
		return
	}
	if weapon == "flag" {
		err = errors.New("'flag' cannot be assigned during re-election.")
		return
	}
	// загрузка произойдёт, если сервер ждёт её, и этот игрок ещё не загрузил ничего.
	if !r.WeaponReElection.WaitingForIt {
		err = errors.New("there is no requirement to re-select a weapon at the moment.")
		return
	}
	if role == 0 {
		reassignWeapons.Rotate()
		if !r.WeaponReElection.User0ReElect {
			if r.UserTurnNumber == 0 {
				r.Map[r.WeaponReElection.AttackingCharacter].Weapon = weapon
			} else {
				r.Map[r.WeaponReElection.AttackedCharacter].Weapon = weapon
			}
			r.WeaponReElection.User0ReElect = true
		} else {
			err = errors.New("You have already downloaded the re-selection.")
			return
		}
	} else {
		if !r.WeaponReElection.User1ReElect {
			if r.UserTurnNumber != 0 {
				r.Map[r.WeaponReElection.AttackingCharacter].Weapon = weapon
			} else {
				r.Map[r.WeaponReElection.AttackedCharacter].Weapon = weapon
			}
			r.WeaponReElection.User1ReElect = true
		} else {
			err = errors.New("You have already downloaded the re-selection.")
			return
		}
	}
	if r.WeaponReElection.User0ReElect && r.WeaponReElection.User1ReElect {
		// то мы как будто бы проводим ход снова, как будто бы небыло перевыбора.
		r.WeaponReElection.WaitingForIt = false
		_, err = r.AttemptGoToCellLogic(r.UserTurnNumber, r.WeaponReElection.AttackingCharacter, r.WeaponReElection.AttackedCharacter)
		if err != nil {
			// Тут точно не должно быть ошибки, которую можно обработать кодом.
			fmt.Print(r.Map)
			panic(err)
		}
	}
	return
}

// ответственность: сформировать изменение для клиента, не изменяет карту.
// считает, что карта уже изменена. Вращает для нулевого игрока.
func (r *Room) MoveCharacter(role RoleId, from int, to int) {
	moveCharacter := types.MoveCharacter{
		From: from,
		To:   to,
	}
	if role == 0 {
		moveCharacter.Rotate()
	}
	response, _ := moveCharacter.MarshalJSON()
	response, _ = types.Event{
		Method:    "move_character",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту.
// считает, что карта уже изменена, поэтому не берёт ничего оттуда, там nil в качестве проигравшего.
// Вращает для нулевого игрока.
func (r *Room) Attack(role RoleId, winner int, winnerWeapon Weapon, loser int, loserWeapon Weapon) {
	attack := types.Attack{
		Winner: types.AttackingСharacter{
			Coordinates: winner,
			Weapon:      string(winnerWeapon),
		},
		Loser: types.AttackingСharacter{
			Coordinates: loser,
			Weapon:      string(loserWeapon),
		},
	}
	if role == 0 {
		attack.Rotate()
	}
	response, _ := attack.MarshalJSON()
	response, _ = types.Event{
		Method:    "attack",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту.
// считает, что карта уже изменена. вращает для нулевого
func (r *Room) AddWeapon(role RoleId, coordinates int, weapon Weapon) {
	addWeapon := types.AddWeapon{
		Coordinates: coordinates,
		Weapon:      string(weapon),
	}
	if role == 0 {
		addWeapon.Rotate()
	}
	response, _ := addWeapon.MarshalJSON()
	response, _ = types.Event{
		Method:    "add_weapon",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: отправка запроса на перевыбор клиенту, не изменяет карту и состояния.
func (r *Room) WeaponChangeRequest(role RoleId, characterOfPlayer int) {
	weaponChangeRequest := types.WeaponChangeRequest{
		CharacterPosition: characterOfPlayer,
	}
	if role == 0 {
		weaponChangeRequest.Rotate()
	}
	response, _ := weaponChangeRequest.MarshalJSON()
	response, _ = types.Event{
		Method:    "weapon_change_request",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту и не прекращает игру.
// считает, что карта уже изменена.
func (r *Room) Gameover(role RoleId, winnerRole RoleId, from int, to int) {
	gameover := types.GameOver{
		Winner: role == winnerRole,
		From:   from,
		To:     to,
	}
	if role == 0 {
		gameover.Rotate()
	}
	response, _ := gameover.MarshalJSON()
	response, _ = types.Event{
		Method:    "gameover",
		Parameter: response,
	}.MarshalJSON()
	if role == 0 {
		r.Messaging.User0To <- response
	} else {
		r.Messaging.User1To <- response
	}
	return
}

// проблемы, почему не используются библиотеки:
// Stateful сервер: необходимо помнить роль, в которой работает пользователь,
// комнату, в которой присутствует пользователь.
// решено делать всё на событиях - клиет пересылает действия пользвателя,
// сервер декларативно присылает изменения, в такой форме, что бы они прямо вызывали анимации.

// сервер получает из одного из двух каналов запись.
// добавляет номер игрока.
// парсит первый уровень.
// находит функцию вызываемую и к ней привязаный тип.
// разворачивает в этот тип пришедшие данные.
