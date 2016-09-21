package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	goquery "github.com/PuerkitoBio/goquery"
)

func checkError(err error, exit bool) {
	if err != nil {
		log.Printf("%+v\n", err)
		if exit {
			os.Exit(-1)
		}
	}
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	checkError(err, true)
	return dir
}

type Currency struct {
	Name       string
	Source     string
	ValFromRMB float64
}

type Config struct {
	Currencys    []Currency
	ToEmails     []string
	SMTPServer   string
	SMTPMail     string
	SMTPPassword string
}

var (
	cfg     Config
	err     error
	isEmail = regexp.MustCompile(`^(\w)+(\.\w+)*@(\w)+((\.\w{2,3}){1,3})$`)
)

func main() {
	mailBody := time.Now().Format("060102")

	cfgFile, err := os.Open(getCurrentDirectory() + `/config.json`)
	checkError(err, true)
	defer cfgFile.Close()
	cfgJSON, _ := ioutil.ReadAll(cfgFile)
	json.Unmarshal(cfgJSON, &cfg)
	//fmt.Println(cfg)

	log.Println("Currencys:")
	for idx := range cfg.Currencys {
		fmt.Println(cfg.Currencys[idx].Name + " | " + cfg.Currencys[idx].Source)
		docEachCurrency, err := goquery.NewDocument(cfg.Currencys[idx].Source)
		checkError(err, true)
		selEachCurrency := docEachCurrency.Find(`div.rate`).First()
		//fmt.Println(selEachCurrency.Text())
		eachCurrency := regexp.MustCompile(`\s`).ReplaceAllString(selEachCurrency.Text(), ``)
		if matched, _ := regexp.MatchString(`\d+\.\d+\/\d+\.\d+.*`, eachCurrency); matched {
			fmt.Println(`####`, `match the format`, ":   ", eachCurrency)
			eachCurrency = regexp.MustCompile(`.*\/(\d+\.\d+).*`).ReplaceAllString(eachCurrency, `$1`)
			f, err := strconv.ParseFloat(eachCurrency, 32)
			checkError(err, true)
			cfg.Currencys[idx].ValFromRMB = f
		} else {
			fmt.Println(`#not`, `match the format`, ":   ", eachCurrency)
			eachCurrency = regexp.MustCompile(`.*(\d+\.\d+).*`).ReplaceAllString(eachCurrency, `$1`)
			f, err := strconv.ParseFloat(eachCurrency, 32)
			checkError(err, true)
			cfg.Currencys[idx].ValFromRMB = 1.0 / f
		}
		fmt.Println(eachCurrency)
	}

	for idx := range cfg.Currencys {
		mailBody += fmt.Sprintf("\n%s: %.4f", cfg.Currencys[idx].Name, cfg.Currencys[idx].ValFromRMB)
	}
	log.Println("MailBody:", "\n", mailBody)

	mails := make(map[string][]string)

	for _, email := range cfg.ToEmails {
		if !isEmail.MatchString(email) {
			log.Println("Wrong Email", email)
			continue
		}
		domain := strings.Split(email, "@")[1]
		mails[domain] = append(mails[domain], email)
	}

	log.Println("Sending Emails:")
	for _, v := range mails {
		for i := 0; i < 5; i++ {
			err = sendToMail(
				cfg.SMTPMail,
				"银联汇率", /*fromName*/
				cfg.SMTPPassword,
				cfg.SMTPServer,
				v,
				"常见币种", /*subject*/
				mailBody)
			if err == nil {
				log.Println("Send mail success!", " | ", v)
				break
			} else {
				checkError(err, false)
				log.Println("Send mail fail! Retry ", i, " | ", v)
				time.Sleep(5 * time.Second)
			}
			time.Sleep(2 * time.Second)
		}
		checkError(err, false)
	}
	log.Println("Send Emails Done")
}
