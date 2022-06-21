package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"main/model"
	"main/repository"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	var argString int
	flag.IntVar(&argString, "port", 8081, "port param")
	flag.Parse()
	port := ":" + strconv.Itoa(argString)

	srv := &http.Server{
		Addr:    port,
		Handler: newRouter(),
	}

	srvCtx, srvStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		shutdownCtx, cancel := context.WithTimeout(srvCtx, 10*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("shutdown time out")
			}
		}()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		srvStopCtx()
	}()

	fmt.Println("Server is running")

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Fatal http server failed: %v", err)
	}

	<-srvCtx.Done()
	fmt.Println("Service stop")
}

func newRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	repository := repository.NewMemory()
	handler := newHandler(repository)

	router.Post("/{proxy}/create", handler.createUser)
	router.Post("/{proxy}/make_friends", handler.makeFriends)
	router.Delete("/{proxy}/user", handler.deleteUser)
	router.Get("/{proxy}/friends/{user_id}", handler.getFriends)
	router.Put("/{proxy}/{user_id}", handler.ageUpdate)

	return router
}

type handler struct {
	repository repository.Storage
}

func newHandler(repository repository.Storage) *handler {
	return &handler{
		repository: repository,
	}
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	json.NewDecoder(r.Body).Decode(user)
	userId, err := h.repository.Create(user)
	if err != nil {
		buildResponse(w, http.StatusInternalServerError, nil)
	}
	defer r.Body.Close()

	result := map[string]int{"id": userId}
	respons, _ := json.Marshal(result)
	buildResponse(w, http.StatusCreated, respons)

}

func (h *handler) makeFriends(w http.ResponseWriter, r *http.Request) {
	makeFriend := &model.MakeFriends{}
	json.NewDecoder(r.Body).Decode(makeFriend)
	defer r.Body.Close()
	sourceName, targetName, err := h.repository.MakeFriends(makeFriend.Source, makeFriend.Target)
	if err != nil {
		log.Print("Failed to make friends users", err)
		buildResponse(w, http.StatusInternalServerError, nil)
		return
	}
	result := sourceName + " и " + targetName + " теперь друзья"
	response, _ := json.Marshal(result)
	buildResponse(w, http.StatusOK, response)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	deleteUser := &model.DeleteUser{}
	json.NewDecoder(r.Body).Decode(deleteUser)
	defer r.Body.Close()
	targetUser, err := h.repository.DeleteUserFromStore(deleteUser.Target)
	if err != nil {
		log.Print("Failed to delete user from store", err)
		buildResponse(w, http.StatusInternalServerError, nil)
		return
	}
	result := targetUser + " удален"
	response, _ := json.Marshal(result)
	buildResponse(w, http.StatusOK, response)
}

func (h *handler) getFriends(w http.ResponseWriter, r *http.Request) {
	stringId := chi.URLParam(r, "user_id")
	userId, err := strconv.Atoi(stringId)
	if err != nil {
		log.Print("Failed conversion stringId to int", err)
		buildResponse(w, http.StatusInternalServerError, nil)
	}
	allFriends, err := h.repository.FriendsReturn(userId)
	if err != nil {
		log.Print("Failed get friends from user", err)
		buildResponse(w, http.StatusInternalServerError, nil)
		return
	}
	result := "друзья: " + allFriends
	response, _ := json.Marshal(result)
	buildResponse(w, http.StatusOK, response)
}

func (h *handler) ageUpdate(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "user_id")
	newAge := &model.AgeUpdate{}
	json.NewDecoder(r.Body).Decode(newAge)
	defer r.Body.Close()
	userId, err := strconv.Atoi(idString)
	if err != nil {
		log.Print("Failed conversion idString to int", err)
		buildResponse(w, http.StatusInternalServerError, nil)
		return
	}
	err = h.repository.AgeUpdate(userId, newAge.Age)
	if err != nil {
		log.Print("Failed to set new age for user", err)
		buildResponse(w, http.StatusInternalServerError, nil)
		return
	}

	result := "Возраст пользователя успешно обновлен"
	response, _ := json.Marshal(result)
	buildResponse(w, http.StatusOK, response)
}

func buildResponse(w http.ResponseWriter, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(body)
}
