package repository

import (
	"context"
	"fmt"
	"log"
	"main/model"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type memory struct {
	sync.Mutex
	index  int
	client *firestore.Client
}

func NewMemory() *memory {
	return &memory{}
}

func (r *memory) AgeUpdate(targetId int, newAge int) error {
	r.Lock()
	defer r.Unlock()
	client, ctx := fbClient()
	defer client.Close()
	docSnaps, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to open collection", err)
		return err
	}
	var idTargetFromFS string
	for _, docSnap := range docSnaps {
		userFb := docSnap.Data()
		idFb := userFb["Id"]
		idInt := int(idFb.(int64))
		if targetId == idInt {
			idTargetFromFS = docSnap.Ref.ID
		}
	}

	_, err = client.Collection("Test").Doc(idTargetFromFS).Update(ctx, []firestore.Update{
		{Path: "Age", Value: newAge},
	})
	if err != nil {
		log.Print("Failed to update collection", err)
		return err
	}
	return nil
}

func (r *memory) FriendsReturn(targetId int) (string, error) {
	r.Lock()
	defer r.Unlock()
	client, ctx := fbClient()
	defer client.Close()
	docSnaps, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to open collection", err)
		return "", err
	}
	var idTargetFromFS string
	for _, docSnap := range docSnaps {
		userFb := docSnap.Data()
		idFb := userFb["Id"]
		idInt := int(idFb.(int64))
		if targetId == idInt {
			idTargetFromFS = docSnap.Ref.ID
		}
	}
	docs, err := client.Collection("Test").Doc(idTargetFromFS).Get(ctx)
	if err != nil {
		log.Print("Failed to open doc", err)
		return "", err
	}
	friendsArray, err := docs.DataAt("Friends")
	if err != nil {
		log.Print("Failed to extracting data from doc", err)
		return "", err
	}
	result := fmt.Sprint(friendsArray)

	return result, nil
}

func (r *memory) DeleteUserFromStore(targetId int) (string, error) {
	client, ctx := fbClient()
	defer client.Close()
	docSnaps, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to open collection", err)
		return "", err
	}
	var idTargetFromFS string
	var targetName string
	for _, docSnap := range docSnaps {
		userFb := docSnap.Data()
		idFb := userFb["Id"]
		idInt := int(idFb.(int64))
		if targetId == idInt {
			idTargetFromFS = docSnap.Ref.ID
			nameFb := userFb["Name"]
			nameStr := nameFb.(string)
			targetName = nameStr
		}
	}
	docs, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to get docs", err)
		return "", err
	}
	for _, snaps := range docs {
		_, err := client.Collection("Test").Doc(snaps.Ref.ID).Update(ctx, []firestore.Update{
			{Path: "Friends", Value: firestore.ArrayRemove(targetName)},
		})
		if err != nil {
			log.Print("Failed to delete user from friens array", err)
			return "", err
		}
	}

	_, err = client.Collection("Test").Doc(idTargetFromFS).Delete(ctx)
	if err != nil {
		log.Print("Failed to delete user", err)
		return "", err
	}

	return targetName, nil
}

func (r *memory) MakeFriends(sourceId, targetId int) (string, string, error) {
	client, ctx := fbClient()
	defer client.Close()
	var sourceUser string
	var targetUser string
	var idSourceFromFS string
	var idTargetFromFS string
	docSnaps, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to open collection", err)
		return "", "", err
	}

	for _, docSnap := range docSnaps {
		userFb := docSnap.Data()
		idFb := userFb["Id"]
		idInt := int(idFb.(int64))
		if sourceId == idInt {
			idSourceFromFS = docSnap.Ref.ID
			nameFb := userFb["Name"]
			nameStr := nameFb.(string)
			sourceUser = nameStr
		} else if targetId == idInt {
			idTargetFromFS = docSnap.Ref.ID
			nameFb := userFb["Name"]
			nameStr := nameFb.(string)
			targetUser = nameStr
		}
	}
	_, err = client.Collection("Test").Doc(idSourceFromFS).Update(ctx, []firestore.Update{
		{Path: "Friends", Value: firestore.ArrayUnion(targetUser)},
	})
	if err != nil {
		log.Print("Failed to update frinds aray in source user", err)
		return "", "", err
	}
	_, err = client.Collection("Test").Doc(idTargetFromFS).Update(ctx, []firestore.Update{
		{Path: "Friends", Value: firestore.ArrayUnion(sourceUser)},
	})
	if err != nil {
		log.Print("Failed to update frinds aray in target user", err)
		return "", "", err
	}
	return sourceUser, targetUser, nil
}

func (r *memory) Create(user *model.User) (int, error) {
	r.Lock()
	defer r.Unlock()
	client, ctx := fbClient()
	defer client.Close()
	docSnaps, err := client.Collection("Test").Documents(ctx).GetAll()
	if err != nil {
		log.Print("Failed to open collection", err)
		return 0, err
	}

	for _, docSnap := range docSnaps {
		userFb := docSnap.Data()
		idFb := userFb["Id"]
		idInt := int(idFb.(int64))
		if r.index < idInt {
			r.index = idInt
		}
	}

	r.index++

	fmt.Println("save index: ", r.index)
	_, _, err = client.Collection("Test").Add(ctx, model.User{
		Id:   r.index,
		Name: user.Name,
		Age:  user.Age,
	})
	if err != nil {
		log.Print("Failed to create doc", err)
		return 0, err
	}

	id := r.index

	return id, nil
}

func fbClient() (client *firestore.Client, ctx context.Context) {
	ctx = context.Background()
	client, err := firestore.NewClient(ctx, "gostudy-ec568", option.WithCredentialsFile("./ServiceAccountKey.json"))
	if err != nil {
		log.Fatal(err)
	}

	return client, ctx
}
