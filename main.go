package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	// database w/ migrations
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// http server
	"github.com/gin-gonic/gin"
)

var conn *sql.DB
var sqliteTimefmt string = "2006-01-02 15:04:05"

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
	dbPath = fmt.Sprintf("%s?_journal=WAL&_timeout=1000&_foreign_keys=1", dbPath)
	conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		panic(err)
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(conn, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite3", driver)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
	m.Close()

	return conn
}

func GetFeed(since *time.Time) []FeedItemJson {
	timestamp := since.Format(sqliteTimefmt)
	rows, err := conn.Query(`
		SELECT id, filepath, channels, datetime
		FROM feed
		WHERE datetime > ?
		ORDER BY datetime DESC
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

		datetime, err := time.Parse(sqliteTimefmt, item.Datetime)
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

func getAboutHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", nil)
}

func getUploadHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

func parseTime(t string) (*time.Time, error) {
	if timestamp, err := strconv.ParseInt(t, 10, 0); err == nil {
		t := time.Unix(timestamp, 0)
		return &t, nil
	}

	if rfc3339, err := time.Parse(time.RFC3339, t); err == nil {
		return &rfc3339, nil
	}

	yyyymmddhhmmss, err := time.Parse(sqliteTimefmt, t)
	if err == nil {
		return &yyyymmddhhmmss, nil
	}

	return nil, err
}

func getFeedHandler(c *gin.Context) {
	since, err := parseTime(c.DefaultQuery("since", "0"))
	if err != nil {
		error := fmt.Sprintf("`since` can either be a unix timestamp in seconds, an RFC3339-formatted datetime or a datetime like `%s`", sqliteTimefmt)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": error,
		})
		return
	}
	feed := GetFeed(since)

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
	router.GET("/about", getAboutHandler)
	router.GET("/feed", getFeedHandler)
	router.GET("/upload", getUploadHandler)
	router.POST("/upload", postUploadHandler)

	router.Run(fmt.Sprintf(":%d", port))
}
