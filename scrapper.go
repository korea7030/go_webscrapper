package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Company struct {
	ID           int    `json:"id"`
	Industryid   int    `json:"industry_id"`
	LogoThumburl string `json:"logo_thumb_url"`
	Name         string `json:"name"`
	OmniText     string `json:"omni_text"`
	PrefText     string `json:"pref_text"`
	RevierScore  int    `json:"review_score"`
	URL          string `json:"url"`
}

func getTotalPage(meta interface{}) int {
	var total int
	for key, data := range meta.(map[string]interface{}) {
		if key == "total" {
			total := int(data.(float64))
			return (total / 10) + 1
		}
	}

	return total
}

func getExtractCompany(jobMap interface{}, c chan<- Company) {
	company := Company{}
	out, _ := json.Marshal(jobMap)
	json.Unmarshal(out, &company)
	// fmt.Println(company)
	c <- company
}

func getJobDataByPage(page int, url string, ch chan []Company) {
	var jobs []Company
	c := make(chan Company)
	URL := url + "&page=" + strconv.Itoa(page)
	resp, err := http.Get(URL)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 빈 interface 선언
	var resultMap map[string]map[string]map[string]interface{}
	// response body를 Byte Array로 변경
	body, err := ioutil.ReadAll(resp.Body)
	// Unmarshal 진행
	json.Unmarshal([]byte(body), &resultMap)
	jobMap := resultMap["data"]["search_result"]["jobs"]
	for _, data := range jobMap.([]interface{}) {
		// fmt.Println("page : ", page)
		// fmt.Println(data.(map[string]interface{})["company"])
		// 생성한 channel(c)에 company 정보 지정
		go getExtractCompany(data.(map[string]interface{})["company"], c)

		// 지정한 정보를 channel(c)에서 뽑아서 사용
		job := <-c
		jobs = append(jobs, job)
	}

	// 리스트를 전달받은 array channel(ch)로 전달
	ch <- jobs
}

func jobPlanetScrapper(url string) {
	var jobs []Company
	resp, err := http.Get(url)
	ch := make(chan []Company)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("status code error : %d, %s", resp.StatusCode, resp.Status)
	}

	if err != nil {
		log.Fatal(err)
	}

	// 빈 interface 선언
	var resultMap map[string]map[string]map[string]interface{}
	// response body를 Byte Array로 변경
	body, err := ioutil.ReadAll(resp.Body)
	// Unmarshal 진행
	json.Unmarshal([]byte(body), &resultMap)
	totalPages := getTotalPage(resultMap["data"]["search_result"]["meta"])
	// go getJobDataByPage(1, url, ch)
	for i := 0; i < totalPages; i++ {
		go getJobDataByPage(i, url, ch)
	}

	for i := 0; i < totalPages; i++ {
		extractJobs := <-ch
		// fmt.Println(extractJobs)
		jobs = append(jobs, extractJobs...)
	}

	// csv 저장
	writeCsvJob(jobs)
}

func writeCsvJob(jobs []Company) {
	file, err := os.Create("jobs.csv")
	if err != nil {
		panic(err)
	}

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"ID", "INDUSTRY_ID", "LOGO_THUMB_URL", "NAME", "OMNI_TEXT", "PREF_TEXT", "REVIEW_SCORE", "URL"}

	wErr := w.Write(header)

	if wErr != nil {
		panic(wErr)
	}

	for _, job := range jobs {
		jobRow := []string{strconv.Itoa(job.ID), strconv.Itoa(job.Industryid), job.LogoThumburl, job.Name, job.OmniText, job.PrefText, strconv.Itoa(job.RevierScore), baseURL + job.URL}
		jobErr := w.Write(jobRow)

		if jobErr != nil {
			panic(jobErr)
		}
	}
}
