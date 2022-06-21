package repository

import (
	"main/model"
)

type Storage interface {
	Create(*model.User) (int, error)
	DeleteUserFromStore(int) (string, error)
	MakeFriends(sourceId, targetId int) (string, string, error)
	FriendsReturn(targetId int) (string, error)
	AgeUpdate(targetId int, newAge int) error
}
