package main

const webDir string = "./web"
const DateStyle string = "20060102"
const DateStylePoints string = "02.01.2006"
const DatabaseDefaultFilePath string = "scheduler.db"

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskId struct {
	Id int `json:"id"`
}

type Tasks struct {
	Tasks []Task `json:"tasks"`
}

type Sign struct {
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
