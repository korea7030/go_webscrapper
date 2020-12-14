package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var fileName = "jobs.csv"
var baseURL = "https://www.jobplanet.co.kr"

func scapeHandler(c *gin.Context) {
	// fmt.Println(c.PostForm("query"))
	query := c.PostForm("query")
	apiURL := baseURL + "/api/v3/job/search?q="

	jobPlanetScrapper(apiURL + query)

	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	// r := c.Request
	w := c.Writer
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	io.Copy(w, f)
}

func main() {
	r := gin.Default()
	r.LoadHTMLFiles("./template/index.html")
	r.GET("/scrapper", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"now": time.Date(2017, 07, 01, 0, 0, 0, 0, time.UTC),
		})
	})
	r.POST("/scrape", scapeHandler)
	r.Run(":4000")
}
