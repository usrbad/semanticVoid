package main

import (
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	file1 = "first.json"
	file2 = "second-part-1.json"
	file3 = "second-part-2.json"
	file4 = "third.json"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	bot, err := tgbotapi.NewBotAPI(os.Getenv("tgToken"))
	checkErr(err)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	checkErr(err)

	for newMsg := range updates{
		if newMsg.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID,stringComposer())
		bot.Send(msg)
	}

}

func openFile(fileName string) []string {
	var words []interface{}
	var result []string
	file, err := os.Open(fileName)
	checkErr(err)
	b, err := ioutil.ReadAll(file)
	checkErr(err)
	json.Unmarshal(b, &words)
	for _, val := range words {
		result = append(result, val.(string))
	}
	file.Close()
	return result
}

func stringComposer() string{
	f1 := openFile(file1)
	f2 := openFile(file2)
	f3 := openFile(file3)
	f4 := openFile(file4)
	return fmt.Sprintf("%v %v%v %v", strings.Title(f1[rand.Intn(len(f1))]), strings.Title(f2[rand.Intn(len(f2))]), f3[rand.Intn(len(f3))], strings.Title(f4[rand.Intn(len(f4))]))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
