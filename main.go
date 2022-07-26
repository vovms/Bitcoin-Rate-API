package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

var emailsMap = map[string]bool{}
var bdFileName = "emails.txt"

func main() {

	initialization()

	basePath := "/api"

	http.HandleFunc(basePath+"/", defaultHandler)
	http.HandleFunc(basePath+"/rate", rateHandler)
	http.HandleFunc(basePath+"/subscribe", subscribeHandler)
	http.HandleFunc(basePath+"/allemails", allemailsHandler)
	http.HandleFunc(basePath+"/deleteallEmails", deleteallEmailsHandler)
	http.HandleFunc(basePath+"/sendEmails", sendEmailsHandler)

	log.Println("Start HTTP server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("defaultHandler EndPoint")
	fmt.Fprintf(w, "defaultHandler")
}

func rateHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("rateHandler EndPoint")

	fmt.Fprintln(w, getCurrentRate())
}

func getCurrentRate() int {
	url := "https://api.apilayer.com/exchangerates_data/convert?to=UAH&from=BTC&amount=1"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", "azhjGjqJCA5ZQby20BPuyZKnQ0t89BXv")

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)

	type resutlStruct struct {
		Result float64 `json:"result"`
	}

	var resutlObj resutlStruct

	json.Unmarshal(body, &resutlObj)

	return int(resutlObj.Result)
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()

	if err != nil {
		log.Fatal("Fatal Error")
	}

	email := r.Form.Get("email")

	if !emailsMap[email] {
		fmt.Fprintln(w, "E-mail додано")
		emailsMap[email] = true
		defer writeEmailToFile(email)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "E-mail вже є в базі даних")
	}
}

func allemailsHandler(w http.ResponseWriter, r *http.Request) {

	for key, _ := range emailsMap {
		fmt.Fprintln(w, key)
	}

}

func deleteallEmailsHandler(w http.ResponseWriter, r *http.Request) {
	emailsMap = map[string]bool{}
	ioutil.WriteFile(bdFileName, []byte(""), 0644)
}

func sendEmailsHandler(w http.ResponseWriter, r *http.Request) {
	sendEmails()
	w.Write([]byte("E-mailʼи відправлено"))
}

func sendEmails() {

	currentRate := getCurrentRate()

	from := "testkulesha@gmail.com"
	password := "snhtioacsnediyzx"
	//toEmailAddress := "kuleshavova@gmail.com"

	mailSlice := make([]string, 0, len(emailsMap))
	for key, _ := range emailsMap {
		mailSlice = append(mailSlice, key)
	}

	to := mailSlice //[]string{toEmailAddress}

	host := "smtp.gmail.com"
	port := "587"
	address := host + ":" + port

	subject := "Subscription to the Bitcoin rate\n"
	body := "Current Rate : " + strconv.Itoa(currentRate)

	message := []byte(subject + body)

	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(address, auth, from, to, message)
	if err != nil {
		panic(err)
	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeEmailToFile(email string) {
	f, err := os.OpenFile(bdFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	email += ";"

	if _, err = f.WriteString(email); err != nil {
		panic(err)
	}
}

func initialization() {
	if !fileExists(bdFileName) {
		f, err := os.Create(bdFileName)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	} else {
		buf, err := ioutil.ReadFile(bdFileName)
		if err != nil {
			panic(err)
		}

		for _, str := range strings.Split(string(buf), ";") {
			if str == "" {
				continue
			}
			emailsMap[str] = true
		}
		//log.Println(emailsMap)
	}
}
