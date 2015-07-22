package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
)

//var Mapping map[string]string = make(map[string]string)
var DB *sql.DB

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	switch {
	case r.Method == "POST":
		name, url := r.Form["name"][0], r.Form["url"][0]
		//Mapping[name] = url
		add_url(DB, name, url)
		http.Redirect(w, r, url, 302)
	case r.URL.Path == "/":
		name := r.Form["name"]
		if len(name) > 0 {
			http.Redirect(w, r, "/"+name[0], 302)
		} else {
			t, _ := template.ParseFiles("index.html")
			context := make(map[string]string)
			t.Execute(w, context)
		}
	case r.URL.Path == "/list":
		mapping := make(map[string]string)
		list_urls(DB, &mapping)
		for name, url := range mapping {
			fmt.Fprintf(w, "%s: %s", name, url)
		}
	default:
		name := r.URL.Path[1:]
		//url := Mapping[name]
		url, err := get_url(DB, name)
		if err != nil {
			fmt.Fprintf(w, "Name not found")
		} else {
			http.Redirect(w, r, url, 302)
		}
	}
}

//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
/*
id int auto_incremental primary key
name varchar(255) unique: lower case
url text
updated datetime auto update

| urls  | CREATE TABLE `urls` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `url` text,
  `updated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 |
*/

type DatabaseConfig struct {
	Type, Name, Host, User, Pass string
	Port                         int
}
type Config struct {
	Database DatabaseConfig
}

func read_conf(filename string) (conf Config) {
	file, err := os.Open(filename)
	dec := json.NewDecoder(file)
	err = dec.Decode(&conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func connect() *sql.DB {
	conf := read_conf("conf.json")
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		conf.Database.User,
		conf.Database.Pass,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Name)

	db, err := sql.Open(conf.Database.Type, conn)
	if err != nil {
		panic(err)
	}
	return db
}

func list_urls(db *sql.DB, urls *map[string]string) {
	rows, err := db.Query("select name, url from urls")
	if err != nil {
		panic(err)
	}
	var name, url string
	for rows.Next() {
		if err := rows.Scan(&name, &url); err != nil {
			panic(err)
		}
		(*urls)[name] = url
	}
}

func get_url(db *sql.DB, name string) (url string, err error) {
	rows, err := db.Query("select url from urls where name = ?", name)
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		if err := rows.Scan(&url); err != nil {
			panic(err)
		}
	} else {
		err = errors.New("Not found")
	}
	return url, err
}

func add_url(db *sql.DB, name, url string) {
	_, err := get_url(db, name)
	fmt.Println(err)
	if err != nil {
		db.Exec("insert into urls(name, url) values(?, ?)", name, url)
	} else {
		db.Exec("update urls set url = ? where name = ?", url, name)
	}
}

func main() {
	DB = connect()
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
