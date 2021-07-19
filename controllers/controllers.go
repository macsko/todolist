package controllers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/kamva/mgm/v3"
	"github.com/macsko/todolist/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

var listModel = &models.List{}
var taskModel = &models.Task{}
var userModel = &models.User{}

// Get JSON request
func getJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return errors.New("decode error")
	}
	return nil
}

// Send JSON response
func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendSuccess(w http.ResponseWriter) {
	sendJSON(w, struct {
		Message string `json:"message"`
	}{
		"success",
	})
}

// Check if user have access to specified list and return this list
func getUserList(w http.ResponseWriter, r *http.Request, listID string) (*models.List, error){
	var list models.List
	err := mgm.Coll(listModel).FindByID(listID, &list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	if list.Username != mux.Vars(r)["username"] {
		http.Error(w, "forbidden resource", http.StatusForbidden)
		return nil, errors.New("forbidden resource")
	}
	return &list, nil
}

// Get user lists
func GetLists(w http.ResponseWriter, r *http.Request) {
	var lists []models.List
	err := mgm.Coll(listModel).SimpleFind(&lists, bson.M{"username": mux.Vars(r)["username"]})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lists == nil {
		lists = []models.List{}
	}
	sendJSON(w, &lists)
}

// Create new user's list
func CreateList(w http.ResponseWriter, r *http.Request) {
	var list models.List
	if err := getJSON(w, r, &list); err != nil {
		return
	}

	list.Username = mux.Vars(r)["username"]
	list.Tasks = []models.Task{}

	err := mgm.Coll(listModel).Create(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, &list)
}

// Update one of user's lists
func UpdateList(w http.ResponseWriter, r *http.Request) {
	var newList models.List
	if err := getJSON(w, r, &newList); err != nil {
		return
	}

	updatedID := mux.Vars(r)["listid"]

	list, err := getUserList(w, r, updatedID)
	if err != nil {
		return // Error handled in previous function
	}

	// Updating fields of List
	list.Title = newList.Title
	if newList.Tasks != nil {
		list.Tasks = newList.Tasks
	}

	err = mgm.Coll(listModel).Update(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, &list)
}

// Delete one of user's lists
func DeleteList(w http.ResponseWriter, r *http.Request) {
	removedID := mux.Vars(r)["listid"]

	removedList, err := getUserList(w, r, removedID)
	if err != nil {
		return // Error handled in previous function
	}

 	for _, task := range removedList.Tasks {
 		err = mgm.Coll(taskModel).Delete(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = mgm.Coll(listModel).Delete(removedList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w)
}

// Get specified list
func GetList(w http.ResponseWriter, r *http.Request) {
	listID := mux.Vars(r)["listid"]

	list, err := getUserList(w, r, listID)
	if err != nil {
		return // Error handled in previous function
	}

	sendJSON(w, &list)
}

// Get specified list of tasks
func GetTasks(w http.ResponseWriter, r *http.Request) {
	listID := mux.Vars(r)["listid"]

	list, err := getUserList(w, r, listID)
	if err != nil {
		return // Error handled in previous function
	}

	sendJSON(w, &list.Tasks)
}

// Get user task from specified list
func GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskid"]
	listID := mux.Vars(r)["listid"]

	var task models.Task
	err := mgm.Coll(taskModel).FindByID(taskID, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = getUserList(w, r, listID) // for authentication
	if err != nil {
		return // Error handled in previous function
	}

	sendJSON(w, &task)
}

// Create new task in user's list
func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := getJSON(w, r, &task); err != nil {
		return
	}

	listID := mux.Vars(r)["listid"]
	list, err := getUserList(w, r, listID)
	if err != nil {
		return // Error handled in previous function
	}

	err = mgm.Coll(taskModel).Create(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list.Tasks = append(list.Tasks, task)
	err = mgm.Coll(listModel).Update(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, &task)
}

// Update one of user's tasks in specified list
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var newTask models.Task
	if err := getJSON(w, r, &newTask); err != nil {
		return
	}

	updatedID := mux.Vars(r)["taskid"]
	listID := mux.Vars(r)["listid"]

	list, err := getUserList(w, r, listID)
	if err != nil {
		return // Error handled in previous function
	}

	var task models.Task
	err = mgm.Coll(taskModel).FindByID(updatedID, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Updating fields of Task
	task.Title = newTask.Title
	task.Body = newTask.Body
	task.Status = newTask.Status

	err = mgm.Coll(taskModel).Update(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, listTask := range list.Tasks {
		if taskID, _ := primitive.ObjectIDFromHex(updatedID); listTask.ID == taskID {
			list.Tasks[i] = task
			err = mgm.Coll(listModel).Update(list)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sendJSON(w, &task)
			return
		}
	}

	http.Error(w, "specified task not found", http.StatusUnprocessableEntity)
}

// Delete one of user's tasks
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	removedID := mux.Vars(r)["taskid"]
	listID := mux.Vars(r)["listid"]

	var removedTask models.Task
	err := mgm.Coll(taskModel).FindByID(removedID, &removedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := getUserList(w, r, listID)
	if err != nil {
		return // Error handled in previous function
	}

	err = mgm.Coll(taskModel).Delete(&removedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, task := range list.Tasks {
		if taskID, _ := primitive.ObjectIDFromHex(removedID); task.ID == taskID {
			list.Tasks = append(list.Tasks[:i], list.Tasks[i+1:]...)
			err = mgm.Coll(listModel).Update(list)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sendSuccess(w)
			return
		}
	}

	http.Error(w, "specified task not found", http.StatusUnprocessableEntity)
}
