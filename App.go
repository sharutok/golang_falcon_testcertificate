package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type qrCode struct {
	Batch_no string `json:"batch_no"`
}
type Result struct {
	FilePath string `json:"file_path"`
}

type Output struct {
	Result Result `json:"result"`
}

func main() {
	var err error
	err = godotenv.Load(".env")
	check("error in loading .env", err)

	db, err = sql.Open("mysql", "Ador_Falcon:S!a?1*_:}aq:(X@tcp(adorwelding.org)/test_certificate?charset=utf8")
	check("error while connection", err)

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/falcon/api/v1", servePDFLink)

	err = http.ListenAndServe(":8080", nil)
	check("error in listening server", err)
}

func servePDFLink(res http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Println("Recovered from panic:", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	var pdf_link string
	api_key := req.Header.Get("api_key")

	err := req.ParseForm()
	httpError(res, "error in parsing form\n", err)

	body, err := io.ReadAll(req.Body)
	httpError(res, "error in reading body\n", err)

	var requestBody qrCode
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		http.Error(res, "Failed to parse JSON body\n", http.StatusBadRequest)
		return
	}
	if requestBody.Batch_no != "" && os.Getenv("API_KEY") == api_key {

		res.Header().Set("Content-Type", "application/json")

		query := fmt.Sprintf("SELECT file_path from test_certificates WHERE batch_no = %v", requestBody.Batch_no)

		err = db.QueryRow(query).Scan(&pdf_link)
		httpError(res, "error in querying data\n", err)

		data := Output{
			Result: Result{
				FilePath: pdf_link,
			},
		}

		json, err := json.Marshal(data)
		httpError(res, "error in marshalling\n", err)

		log.Println("Api called!")
		res.Write(json)
	} else {
		fmt.Fprintln(res, "batch no does not exisit")
	}

}

func check(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
		return
	}
}

func httpError(resp http.ResponseWriter, mes string, err error) {
	if err != nil {
		http.Error(resp, "Failed to read request body", http.StatusInternalServerError)
		return
	}
}
