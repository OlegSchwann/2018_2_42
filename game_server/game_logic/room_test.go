package game_logic

import (
	"github.com/OlegSchwann/rpsarena-ru-backend/game_server/types"
	"reflect"
	"testing"
)

// Тестирование алгоритма формирования карты для пользователя 0.
func Test_uploadMapRole0(t *testing.T) {
	r := Room{}
	role := RoleId(0)
	uploadedMap := types.UploadMap{Weapons: [14]string{
		"rock", "scissors", "paper", "rock", "scissors", "paper", "rock",
		"scissors", "paper", "rock", "flag", "scissors", "paper", "rock",
	}}

	Expected := Map{
		&Сharacter{
			Role:         0,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "flag",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         0,
			Weapon:       "rock",
			ShowedWeapon: false,
		},
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
	}

	_ = r.uploadMap(role, uploadedMap)
	if !reflect.DeepEqual(Expected, r.Map) {
		t.Errorf("RoomUploadMap for Role 0, expected\n%s\ngot\n%s\n", Expected.String(), r.Map.String())
	}
}

// Тестирование алгоритма формирования карты для пользователя 0.
func Test_uploadMapRole1(t *testing.T) {
	r := Room{}
	role := RoleId(1)
	uploadedMap := types.UploadMap{Weapons: [14]string{
		"rock", "paper", "scissors", "flag", "rock", "paper", "scissors",
		"rock", "paper", "scissors", "rock", "paper", "scissors", "rock",
	}}

	Expected := Map{
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil,
		&Сharacter{
			Role:         1,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "flag",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "rock",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "paper",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "scissors",
			ShowedWeapon: false,
		}, &Сharacter{
			Role:         1,
			Weapon:       "rock",
			ShowedWeapon: false,
		},
	}

	_ = r.uploadMap(role, uploadedMap)
	if !reflect.DeepEqual(Expected, r.Map) {
		t.Errorf("RoomUploadMap for Role 0, expected\n%s\ngot\n%s\n", Expected.String(), r.Map.String())
	}
}
