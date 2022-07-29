package main

import (
	"encoding/json"
	"errors"
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
	http.HandleFunc(basePath+"/allemails", allemailsHandler)
	http.HandleFunc(basePath+"/deleteallEmails", deleteallEmailsHandler)

	http.HandleFunc(basePath+"/rate", rateHandler)
	http.HandleFunc(basePath+"/subscribe", subscribeHandler)
	http.HandleFunc(basePath+"/sendEmails", sendEmailsHandler)

	log.Println("Start HTTP server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("defaultHandler EndPoint")
	fmt.Fprintf(w, "defaultHandler")
}

func rateHandler(w http.ResponseWriter, r *http.Request) {

	currentRate, err := getCurrentRate()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Invalid status value")
		return
	}

	fmt.Fprintln(w, currentRate)

}

func getCurrentRate() (int, error) {
	url := "https://api.apilayer.com/exchangerates_data/convert?to=UAH&from=BTC&amount=1"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", "azhjGjqJCA5ZQby20BPuyZKnQ0t89BXv")

	if err != nil {
		log.Println(err)
		return 0, err
	}
	res, erro := client.Do(req)

	if erro != nil {
		log.Println(erro)
		return 0, erro
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	type resutlStruct struct {
		Result float64 `json:"result"`
	}

	var resutlObj resutlStruct

	err = json.Unmarshal(body, &resutlObj)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(resutlObj.Result), nil
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Invalid email value")
		return
	}
	email := r.Form.Get("email")
	if email == "" {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Invalid email value")
		return
	}

	err = subscribe(email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintln(w, "E-mail додано")

}

func subscribe(email string) error {

	if !emailsMap[email] {
		emailsMap[email] = true
		defer writeEmailToFile(email)
		return nil
	} else {
		return errors.New("E-mail вже є в базі даних")
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
	err := sendEmails()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	w.Write([]byte("E-mailʼи відправлено"))
}

func sendEmails() error {

	currentRate, err := getCurrentRate()
	if err != nil {
		return err
	}

	from := "testkulesha@gmail.com"
	password := "snhtioacsnediyzx"

	mailSlice := make([]string, 0, len(emailsMap))
	for key, _ := range emailsMap {
		mailSlice = append(mailSlice, key)
	}

	to := mailSlice

	host := "smtp.gmail.com"
	port := "587"
	address := host + ":" + port

	subject := "Subscription to the Bitcoin rate\n"
	body := "Current Rate : " + strconv.Itoa(currentRate)

	message := []byte(subject + body)

	auth := smtp.PlainAuth("", from, password, host)

	err = smtp.SendMail(address, auth, from, to, message)
	if err != nil {
		return err
	}

	return nil
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
