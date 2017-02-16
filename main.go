package main

import (
	"fmt"
	"net/http"
	"database/sql"
	"os"
	"log"
	"os/signal"
	"syscall"
	"github.com/miekg/dns"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("location","/ws")
}
const (
	None = iota
	MySQL
	Postgres
	SQLite
)
var driver = map[int]string{
	MySQL: "mysql",
	Postgres: "postgres",
	SQLite: "sqlite3",
}
var configDriver = map[string]int{
	"mysql": MySQL,
	"postgres": Postgres,
	"sqlite": SQLite,
}
var CREATEUSERS = map[int]string{
	MySQL: "CREATE TABLE IF NOT EXISTS users (id INT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), key VARCHAR(16), port INT UNSIGNED NOT NULL, PRIMARY KEY (id))",
	Postgres: "CREATE SEQUENCE users_seq; CREATE TABLE IF NOT EXISTS users (id INT CHECK (id > 0) NOT NULL DEFAULT NEXTVAL ('users_seq'), name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), port INT CHECK (port > 0) NOT NULL, PRIMARY KEY (id)) create index VARCHAR on users(16);",
	SQLite: "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), key VARCHAR(16), port INTEGER UNSIGNED NOT NULL)",
}
var CREATEMESSAGES = map[int]string{
	MySQL: "CREATE TABLE IF NOT EXISTS messages (`id` INT UNSIGNED NOT NULL AUTO_INCREMENT, `to` INT UNSIGNED NOT NULL, `from` INT UNSIGNED NOT NULL, `time` INT UNSIGNED NOT NULL, `type` VARCHAR(255), `message` longtext NULL, PRIMARY KEY (id))",
	Postgres: "CREATE SEQUENCE messages_seq; CREATE TABLE IF NOT EXISTS messages (id INT CHECK (id > 0) NOT NULL DEFAULT NEXTVAL ('messages_seq'), to INT CHECK (to > 0) NOT NULL, from INT CHECK (from > 0) NOT NULL, time INT CHECK (time > 0) NOT NULL, type VARCHAR(255), message longtext NULL, PRIMARY KEY (id))",
	SQLite: "CREATE TABLE IF NOT EXISTS messages (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `to` INTEGER UNSIGNED NOT NULL, `from` INTEGER UNSIGNED NOT NULL, `time` INTEGER UNSIGNED NOT NULL, `type` VARCHAR(255), `message` longtext NULL)",
}
func main(){
	viper.SetConfigName("gpm")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/etc/gpm/")
	viper.AddConfigPath("$HOME/.gpm")
	viper.AddConfigPath(".")
	viper.SetDefault("port",8080)
	viper.SetDefault("dns.enabled",true)
	viper.SetDefault("dns.port",5353)
	viper.SetDefault("database.Type","sqlite")
	viper.SetDefault("database.File","./gpm.db")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("No configuration file found - using defaults")
	}

	dbType := configDriver[viper.GetString("database.Type")]
	var source string
	if dbType == SQLite {
		source = viper.GetString("database.File")
	}else if dbType == Postgres {
		source = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",viper.GetString("database.Username"),viper.GetString("database.Password"),viper.GetString("database.Hostname"),viper.GetInt("database.Port"),viper.GetString("database.Database"))
	}else if dbType == MySQL {
		if viper.GetString("database.Socket") != "" {
			source = fmt.Sprintf("%s:%s@unix(%s)/%s",viper.GetString("database.Username"),viper.GetString("database.Password"),viper.GetString("database.Socket"),viper.GetString("database.Database"))
		}else{
			viper.SetDefault("database.Port",3306)
			source = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",viper.GetString("database.Username"),viper.GetString("database.Password"),viper.GetString("database.Hostname"),viper.GetInt("database.Port"),viper.GetString("database.Database"))
		}
	}else{
		log.Fatalf("Unreconized Database Type")
	}
	db,err:=sql.Open(driver[dbType],source)
	if err != nil {
		log.Fatalf("Unable To Connect To Database: %s",err)
		return
	}
	_,err = db.Exec(CREATEUSERS[dbType])
	if err != nil {
		log.Fatalf("Unable To Create Users Table: %s",err)
		return
	}
	_,err = db.Exec(CREATEMESSAGES[dbType])
	if err != nil {
		log.Fatalf("Unable To Create Messages Table: %s",err)
		return
	}
	defer db.Close()
	tracker := newTracker()
	go tracker.run()
	http.HandleFunc("/",serveHome)
	http.HandleFunc("/ws",func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(db,tracker,w,r)
	})
	go func(){
		err := http.ListenAndServe(fmt.Sprintf(":%d",viper.GetInt("port")),nil)
		if err != nil {
			log.Fatalf("Unable To Listen on port %d for websocket: ",viper.GetInt("port"),err)
		}
	}()
	if viper.GetBool("dns.enabled") {
		viper.SetDefault("dns.domain","example.com")
		dns.HandleFunc(".",func(w dns.ResponseWriter, r *dns.Msg){
			serveDNSRequest(db,w,r)
		})
		go func(){
			err := dns.ListenAndServe(fmt.Sprintf(":%d",viper.GetInt("dns.port")),"tcp",nil)
			if err != nil {
				log.Fatalf("Unable To Listen on port %d/tcp for dns server: ",viper.GetInt("dns.port"),err)
			}
		}()
		go func(){
			err := dns.ListenAndServe(fmt.Sprintf(":%d",viper.GetInt("dns.port")),"udp",nil)
			if err != nil {
				log.Fatalf("Unable To Listen on port %d/udp for dns server: ",viper.GetInt("dns.port"),err)
			}
		}()
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Fatalf("Signal (%d) received, stopping\n",s)
		}
	}
}
