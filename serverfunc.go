package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

func NextDateReadGET(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse(DateStyle, r.FormValue("now"))
	if err != nil {
		responseWithError(w, "Ошибка формата даты", err)
		return
	}

	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	nextDate, err := NextDate(now, date, repeat)

	if err != nil {
		responseWithError(w, "Ошибка вычисления правила повторения задачи", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))

	if err != nil {
		log.Println(fmt.Print("w.Write error"))
		responseWithError(w, "Ошибка отправки данных", err)
	}
}

func responseWithError(w http.ResponseWriter, errorText string, err error) {
	errorResponse := ErrorResponse{Error: fmt.Errorf("%s: %w", errorText, err).Error()}
	errorData, _ := json.Marshal(errorResponse)
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write(errorData)

	if err != nil {
		log.Println(fmt.Print("w.Write error"))
		responseWithError(w, "Внутренняя ошибка сервера", err)
	}
}

func TaskAddPOST(w http.ResponseWriter, r *http.Request) {
	var taskData Task
	var buffer bytes.Buffer

	if _, err := buffer.ReadFrom(r.Body); err != nil {
		log.Println(fmt.Print("buffer.ReadFrom(r.Body) error"))
		responseWithError(w, "Ошибка формы запроса", err)
		return
	}

	if err := json.Unmarshal(buffer.Bytes(), &taskData); err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(taskData.Date) == 0 {
		taskData.Date = time.Now().Format(DateStyle)
	} else {
		date, err := time.Parse(DateStyle, taskData.Date)
		if err != nil {
			log.Println(fmt.Print("DateStyle error"))
			responseWithError(w, "Ошибка формата даты", err)
			return
		}

		if date.Before(time.Now()) {
			taskData.Date = time.Now().Format(DateStyle)
		}
	}

	if len(taskData.Title) == 0 {
		log.Println(fmt.Print("Title error"))
		responseWithError(w, "Ошибка загаловка", errors.New("Title"))
		return
	}

	if len(taskData.Repeat) > 0 {
		if _, err := NextDate(time.Now(), taskData.Date, taskData.Repeat); err != nil {
			log.Println(fmt.Print("Repeat error"))
			responseWithError(w, "Ошибка вычисления правила повторения задачи", err)
			return
		}
	}

	taskId, err := InsertTask(taskData)
	if err != nil {
		log.Println(fmt.Print("Insert task error"))
		responseWithError(w, "Ошибка добавления задачи", err)
		return
	}

	taskIdData, err := json.Marshal(TaskId{Id: taskId})
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(taskIdData)

	if err != nil {
		log.Println(fmt.Print("w.Write(taskIdData) error"))
		responseWithError(w, "Ошибка отправки данных", err)
	}
}

func TasksReadGET(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	var tasks []Task

	if len(search) > 0 {
		date, err := time.Parse(DateStylePoints, search)
		if err != nil {
			tasks, err = SearchTasks(search)
			if err != nil {
				log.Println(fmt.Print("SearchTasks error"))
				responseWithError(w, "Ошибка поиска задачи", err)
			}
		} else {
			tasks, err = SearchTasksByDate(date.Format(DateStyle))
			if err != nil {
				log.Println(fmt.Print("SearchTasksByDate error"))
				responseWithError(w, "Ошибка поиска задачи по дате", err)
			}
		}
	} else {
		var err error
		tasks, err = ReadTasks()
		if err != nil {
			log.Println(fmt.Print("SearchTasksByDate error"))
			responseWithError(w, "Ошибка чтения задач", err)
			return
		}
	}

	tasksData, err := json.Marshal(Tasks{Tasks: tasks})
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(tasksData)

	if err != nil {
		responseWithError(w, "Write(tasksData) error", err)
	}
}

func TaskReadGET(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if len(id) == 0 {
		log.Println(fmt.Print("id error"))
		responseWithError(w, "Ошибка идентификатора", errors.New("id"))
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		log.Println(fmt.Print("Atoi(task.Id) error"))
		responseWithError(w, "Ошибка фидентификатора", err)
		return
	}

	task, err := ReadTask(id)
	if err != nil {
		responseWithError(w, "ReadTask(id) error", err)
		return
	}

	tasksData, err := json.Marshal(task)
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(tasksData)

	if err != nil {
		log.Println(fmt.Print("w.Write(taskIdData) error"))
		responseWithError(w, "Ошибка отправки данных", err)
	}
}

func TaskUpdatePUT(w http.ResponseWriter, r *http.Request) {
	var task Task
	var buffer bytes.Buffer

	if _, err := buffer.ReadFrom(r.Body); err != nil {
		log.Println(fmt.Print("buffer.ReadFrom(r.Body) error"))
		responseWithError(w, "Ошибка формы запроса", err)
		return
	}

	if err := json.Unmarshal(buffer.Bytes(), &task); err != nil {
		log.Println(fmt.Print("json.Unmarshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(task.Id) == 0 {
		log.Println(fmt.Print("id error"))
		responseWithError(w, "Ошибка идентификатора", errors.New("id"))
		return
	}

	if _, err := strconv.Atoi(task.Id); err != nil {
		log.Println(fmt.Print("Atoi(task.Id) error"))
		responseWithError(w, "Ошибка фидентификатора", err)
		return
	}

	if _, err := time.Parse(DateStyle, task.Date); err != nil {
		log.Println(fmt.Print("DateStyle error"))
		responseWithError(w, "Ошибка формата даты", err)
		return
	}

	if len(task.Title) == 0 {
		log.Println(fmt.Print("Title error"))
		responseWithError(w, "Ошибка загаловка", errors.New("Title"))
		return
	}

	if len(task.Repeat) > 0 {
		if _, err := NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			responseWithError(w, "Ошибка вычисления правила повторения задачи", err)
			return
		}
	}

	_, err := UpdateTask(task)
	if err != nil {
		log.Println(fmt.Print("UpdateTask(task) error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	taskIdData, err := json.Marshal(task)
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(taskIdData)

	if err != nil {
		log.Println(fmt.Print("w.Write(taskIdData) error"))
		responseWithError(w, "Ошибка отправки данных", err)
		return
	}
}

func TaskDonePOST(w http.ResponseWriter, r *http.Request) {
	task, err := ReadTask(r.URL.Query().Get("id"))
	if err != nil {
		responseWithError(w, "ReadTask error", err)
		return
	}

	if len(task.Repeat) == 0 {
		err = DeleteTask(task.Id)
		if err != nil {
			responseWithError(w, " DeleteTask(task.Id) error", err)
			return
		}
	} else {
		task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			responseWithError(w, "NextDate error", err)
			return
		}

		_, err = UpdateTask(task)
		if err != nil {
			log.Println(fmt.Print("UpdateTask(task) error"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tasksData, err := json.Marshal(struct{}{})
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(tasksData)

	if err != nil {
		log.Println(fmt.Print("w.Write(taskIdData) error"))
		responseWithError(w, "Ошибка отправки данных", err)
	}
}

func TaskDELETE(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := DeleteTask(id)
	if err != nil {
		responseWithError(w, "DeleteTask error", err)
		return
	}

	tasksData, err := json.Marshal(struct{}{})
	if err != nil {
		log.Println(fmt.Print("json.Marshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(tasksData)

	if err != nil {
		log.Println(fmt.Print("w.Write(taskIdData) error"))
		responseWithError(w, "Ошибка отправки данных", err)
		return
	}
}

func SignInPOST(w http.ResponseWriter, r *http.Request) {
	var signData Sign
	var buffer bytes.Buffer

	if _, err := buffer.ReadFrom(r.Body); err != nil {
		log.Println(fmt.Print("buffer.ReadFrom(r.Body) error"))
		responseWithError(w, "Ошибка формы запроса", err)
		return
	}

	if err := json.Unmarshal(buffer.Bytes(), &signData); err != nil {
		log.Println(fmt.Print("json.Unmarshal error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	envPassword := os.Getenv("TODO_PASSWORD")

	//log.Println(fmt.Print("token = ", signData.Password))
	//log.Println(fmt.Print("envPassword = ", envPassword))
	if signData.Password == envPassword {
		jwtInstance := jwt.New(jwt.SigningMethodHS256)
		token, err := jwtInstance.SignedString([]byte(envPassword))
		if err != nil {
			responseWithError(w, "jwtInstance.SignedString error", err)
		}

		taskIdData, err := json.Marshal(Token{Token: token})
		if err != nil {
			log.Println(fmt.Print("json.Marshal error"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(taskIdData)

		if err != nil {
			log.Println(fmt.Print("w.Write(taskIdData) error"))
			responseWithError(w, "Ошибка отправки данных", err)
		}
	} else {
		errorResponse := ErrorResponse{Error: "wrong password"}
		errorData, _ := json.Marshal(errorResponse)
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write(errorData)

		if err != nil {
			log.Println(fmt.Print("w.Write(taskIdData) error"))
			responseWithError(w, "Ошибка отправки данных", err)
		}
	}
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var cookieToken string
			cookie, err := r.Cookie("token")
			if err == nil {
				cookieToken = cookie.Value
			}
			jwtInstance := jwt.New(jwt.SigningMethodHS256)
			token, err := jwtInstance.SignedString([]byte(pass))
			if err != nil {
				responseWithError(w, "jwtInstance.SignedString", err)
			}

			if cookieToken != token {
				errorResponse := ErrorResponse{Error: "wrong password"}
				errorData, _ := json.Marshal(errorResponse)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(errorData)

				if err != nil {
					log.Println(fmt.Print("w.Write(taskIdData) error"))
					responseWithError(w, "Ошибка отправки данных", err)
				}
				return
			}
		}
		next(w, r)

	}

}
