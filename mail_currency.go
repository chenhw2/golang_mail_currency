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
	"time"

	goquery "github.com/PuerkitoBio/goquery"
)

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	checkError(err)
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

func main() {
	mailBody := time.Now().Format("060102")
	var cfg Config
	var err error
	//fmt.Println(getCurrentDirectory())
	cfgFile, err := os.Open(getCurrentDirectory() + `/config.json`)
	checkError(err)
	defer cfgFile.Close()
	cfgJSON, _ := ioutil.ReadAll(cfgFile)
	json.Unmarshal(cfgJSON, &cfg)

	log.Println("\n ** Currencys:")
	for idx := range cfg.Currencys {
		fmt.Println(cfg.Currencys[idx].Name + " | " + cfg.Currencys[idx].Source)
		docEachCurrency, err := goquery.NewDocument(cfg.Currencys[idx].Source)
		checkError(err)
		selEachCurrency := docEachCurrency.Find(`div.rate`).First()
		//fmt.Println(selEachCurrency.Text())
		eachCurrency := regexp.MustCompile(`\s`).
			ReplaceAllString(selEachCurrency.Text(), ``)
		if matched, _ := regexp.MatchString(`\d+\.\d+\/\d+\.\d+.*`, eachCurrency); matched {
			fmt.Println(`####`, `match the format`, ":   ", eachCurrency)
			eachCurrency = regexp.MustCompile(`.*\/(\d+\.\d+).*`).ReplaceAllString(eachCurrency, `$1`)
			f, err := strconv.ParseFloat(eachCurrency, 32)
			checkError(err)
			cfg.Currencys[idx].ValFromRMB = f
		} else {
			fmt.Println(`#not`, `match the format`, ":   ", eachCurrency)
			eachCurrency = regexp.MustCompile(`.*(\d+\.\d+).*`).ReplaceAllString(eachCurrency, `$1`)
			f, err := strconv.ParseFloat(eachCurrency, 32)
			checkError(err)
			cfg.Currencys[idx].ValFromRMB = 1 / f
		}
		fmt.Println(eachCurrency)
	}

	for idx := range cfg.Currencys {
		mailBody += fmt.Sprintf("\n%s: %.4f", cfg.Currencys[idx].Name, cfg.Currencys[idx].ValFromRMB)
	}
	log.Println(" ** mailBody:,", "\n", mailBody)

	log.Println(" ** Sending Emails:")
	for _, email := range cfg.ToEmails {
		log.Print(email)
		continue
		for i := 0; i < 10; i++ {
			err = sendToMail(
				cfg.SMTPMail,     /*fromMail*/
				"银联汇率",           /*fromName*/
				cfg.SMTPPassword, /*password*/
				cfg.SMTPServer,   /*smtpServer*/
				email,
				"",     /*toName*/
				"常见币种", /*subject*/
				mailBody)
			if err == nil {
				log.Println(":", "Send mail success!")
				break
			} else {
				log.Println(":", "Send mail fail! Retry ", i)
			}
		}
		checkError(err)
	}
}
