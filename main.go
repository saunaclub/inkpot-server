package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func getIndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func getUploadHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

func postUploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")

	// The file cannot be received.
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file is received",
		})
		return
	}

	// Retrieve file information
	// extension := filepath.Ext(file.Filename)
	// Generate random file name for the new uploaded file so it doesn't override the old file with same name
	newFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)

	// The file is received, so let's save it
	if err := c.SaveUploadedFile(file, "./uploads/"+newFileName); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to save the file",
		})
		return
	}

	// File saved successfully. Return proper result
	c.JSON(http.StatusOK, gin.H{
		"message": "Your file has been successfully uploaded.",
	})
}

func main() {
	// We want to make the HTTP server port configurable, which we do by
	// defining a `-p` flag
	var port int
	flag.IntVar(&port, "p", 8000, "Port of webserver, defaults to 8000")
	flag.Parse()

	router := gin.Default()

	// Set up template handling; we're working with simple HTML templates for
	// now, which are all in the `templates` directory. Some useful docs:
	// - https://golangdocs.com/templates-in-golang
	router.LoadHTMLGlob("templates/*")

	// Finally, set up the different routes
	router.GET("/", getIndexHandler)
	router.GET("/upload", getUploadHandler)
	router.POST("/upload", postUploadHandler)

	router.Run(fmt.Sprintf(":%d", port))
}
