package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	// database w/ migrations
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	// http server
	"github.com/gin-gonic/gin"
)

var conn *sql.DB

type FeedItemRaw struct {
	Id       int
	Channels string
	FilePath string
	Datetime string
}

type FeedItemJson struct {
	Id       int       `json:"id"`
	Channels []string  `json:"channels"`
	Url      string    `json:"url"`
	Datetime time.Time `json:"datetime"`
}

func setupDb(dbPath string) *sql.DB {
	var err error

	conn, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal=WAL&_timeout=1000&_foreign_keys=1", dbPath))
	if err != nil {
		panic(err)
	}

	initDb := `
    	CREATE TABLE IF NOT EXISTS feed (
			id	INTEGER	PRIMARY	KEY	AUTOINCREMENT	NOT	NULL,
			filepath	TEXT	NOT	NULL,
			channels	TEXT	NOT	NULL,
			datetime	TEXT	NOT	NULL
		)
	`
	if _, err := conn.Exec(initDb, nil); err != nil {
		panic(err)
	}
	return conn
}

func GetFeed(since *time.Time) []FeedItemJson {
	timefmt := "2006-01-02 15:04:05"
	timestamp := since.Format(timefmt)
	rows, err := conn.Query(`
		SELECT id, filepath, channels, datetime
        FROM feed
        WHERE datetime > ?
        ORDER BY datetime
        LIMIT 10
    `, timestamp)
	if err != nil {
		log.Fatalf("Error generating feed: %s", err)
	}
	defer rows.Close()

	var items = make([]FeedItemJson, 0)
	for rows.Next() {
		item := new(FeedItemRaw)
		if err := rows.Scan(&item.Id, &item.FilePath, &item.Channels, &item.Datetime); err != nil {
			log.Fatalf("Error fetching feed row: %s", err)
		}

		datetime, err := time.Parse(timefmt, item.Datetime)
		if err != nil {
			log.Fatalf("Error parsing datetime in row: %s", err)
		}

		json := FeedItemJson{
			Id:       item.Id,
			Channels: strings.Split(item.Channels, ", "),
			Datetime: datetime,
			Url:      item.FilePath,
		}
		items = append(items, json)
	}

	return items
}

func getIndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func getUploadHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

func getFeedHandler(c *gin.Context) {
	loc := time.FixedZone("UTC", 0)
	t := time.Date(1970, time.January, 1, 0, 0, 0, 0, loc)
	feed := GetFeed(&t)

	c.JSON(http.StatusOK, gin.H{
		"data": feed,
	})
}

func postUploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	channels := c.PostForm("channels")

	// The file cannot be received.
	if err != nil {
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"error": "Please select a file to upload.",
		})
		return
	}

	if channels == "" {
		channels = "default"
	}

	// Retrieve file information
	// extension := filepath.Ext(file.Filename)
	// Generate random file name for the new uploaded file so it doesn't override the old file with same name
	newFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	path := "./uploads/" + newFileName

	// The file is received, so let's save it
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"error": "Could not save uploaded file.",
		})
		return
	}

	if _, err := conn.Exec(`
		INSERT INTO feed (filepath, channels, datetime)
		VALUES (?, ?, datetime("now"))
	`, path, channels); err != nil {
		log.Fatalf("Error creating database entry while uploading: %s", err)
	}

	// File saved successfully. Return proper result
	c.HTML(http.StatusOK, "upload.html", gin.H{
		"info": "Thanks! Your image was added to the feed.",
	})
}

func main() {
	// We want to make the HTTP server port configurable, which we do by
	// defining a `-p` flag
	var port int
	flag.IntVar(&port, "p", 8000, "Port of webserver, defaults to 8000")
	flag.Parse()

	// Set up database
	conn = setupDb("inkpot.db")
	defer conn.Close()

	// Start http server
	router := gin.Default()

	// Set up template handling; we're working with simple HTML templates for
	// now, which are all in the `templates` directory. Some useful docs:
	// - https://golangdocs.com/templates-in-golang
	router.LoadHTMLGlob("templates/*")

	router.Static("/assets", "./assets")
	router.Static("/uploads", "./uploads")

	// Finally, set up the different routes
	router.GET("/", getIndexHandler)
	router.GET("/feed", getFeedHandler)
	router.GET("/upload", getUploadHandler)
	router.POST("/upload", postUploadHandler)

	router.Run(fmt.Sprintf(":%d", port))
}
