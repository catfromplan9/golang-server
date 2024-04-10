package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
)

var sql_path = "./sql"

var db *sql.DB

var database_version = "dev2"

func db_create() error {
	schema, err := ioutil.ReadFile(sql_path + "/schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}

	err = db_pair_store("version", database_version)
	if err != nil {
		return err
	}

	return nil
}

func db_init() error {
	var err error
	var ver string

	db, err = sql.Open("sqlite3", sql_path+"/database.db")
	if err != nil {
		return err
	}

	ver, err = db_pair_load("version")
	if err != nil {
		err = db_create()
		if err != nil {
			return err
		}

		ver, err = db_pair_load("version")
		if err != nil {
			return err
		}

		status, err := account_register("Administrator", "admin", "admin@localhost")
		if err != nil {
			return err
		}
		if status != "" {
			return errors.New(status)
		}
	}

	if ver != database_version {
		return errors.New("Bad database version")
	}

	return err
}

func db_pair_load(key string) (string, error) {
	var value string
	row := db.QueryRow("SELECT value FROM variable WHERE key=?1", key)
	err := row.Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func db_pair_store(key string, value string) error {
	_, err := db.Exec("REPLACE INTO variable (key, value) VALUES (?1, ?2)", key, value)
	if err != nil {
		return err
	}
	return nil
}
