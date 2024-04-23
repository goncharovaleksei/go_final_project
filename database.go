package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func getDatabaseFilePath() string {
	envDbFilePath := os.Getenv("TODO_DBFILE")
	if len(envDbFilePath) > 0 {
		return envDbFilePath
	}

	return DatabaseDefaultFilePath
}

func createDatabaseFile(dbFilePath string) (*sql.DB, error) {
	_, err := os.Create(dbFilePath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTable(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS `scheduler` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `date` VARCHAR(8) NULL, `title` VARCHAR(64) NOT NULL, `comment` VARCHAR(255) NULL, `repeat` VARCHAR(128) NULL)")
	if err != nil {
		log.Fatal(err)
	}
}

func DatabaseInit() {
	dbFilePath := getDatabaseFilePath()
	_, err := os.Stat(dbFilePath)

	if err != nil {
		db, err = createDatabaseFile(dbFilePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db, err = sql.Open("sqlite3", dbFilePath)
	}

	if err != nil {
		log.Fatal(err)
	}
	createTable(db)
}

func InsertTask(task Task) (int, error) {
	result, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func ReadTasks() ([]Task, error) {
	var tasks []Task

	rows, err := db.Query("SELECT * FROM scheduler ORDER BY date")
	if err != nil {
		return []Task{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []Task{}, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return []Task{}, err
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

func SearchTasks(search string) ([]Task, error) {
	var tasks []Task

	search = fmt.Sprintf("%%%s%%", search)
	rows, err := db.Query("SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date", sql.Named("search", search))
	if err != nil {
		return []Task{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []Task{}, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return []Task{}, err
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

func SearchTasksByDate(date string) ([]Task, error) {
	var tasks []Task

	rows, err := db.Query("SELECT * FROM scheduler WHERE date = :date", sql.Named("date", date))
	if err != nil {
		return []Task{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []Task{}, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return []Task{}, err
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

func ReadTask(id string) (Task, error) {
	var task Task

	row := db.QueryRow("SELECT * FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		return Task{}, err
	}

	return task, nil
}

func UpdateTask(task Task) (Task, error) {
	result, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return Task{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Task{}, err
	}

	if rowsAffected == 0 {
		return Task{}, errors.New("failed to update")
	}

	return task, nil
}

func DeleteTask(id string) error {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("failed to delete")
	}

	return err
}
