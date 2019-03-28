/*
###################################################
Простой бот, который генерирует фразы из словаря.
Для работы пропишите две вещи:
1. Переменную окружения tgToken (либо укажите прямо в коде)
2. Переменную superadmin - ваш логин в телеге без @
Доступные команды:
updateDB - обновить базу слов.
admins - список админов.
addadmin - добавить админа (проверки существования логина нет, пока не делал, вводите верный логин).
deladmin - удалить админа.
help - справка по командам.
Все команды кроме help работают только от админа.
###################################################
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	file1 = "first.json"
	file2 = "second-part-1.json"
	file3 = "second-part-2.json"
	file4 = "third.json"
)

var files struct {
	f1, f2, f3, f4 []string
}

func main() {

	superadmin := "" //superadmin username
	admins := make(map[string]bool)
	admins[superadmin] = true
	rand.Seed(time.Now().UTC().UnixNano())
	bot, err := tgbotapi.NewBotAPI(os.Getenv("tgToken"))
	checkErr(err)
	readWordsFromFiles()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	checkErr(err)
	// Special for Heroku:
	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	// end special
	lastmsg := ""
	for newMsg := range updates {
		_, ok := admins[newMsg.Message.From.UserName]
		switch {
		case newMsg.Message == nil:
			fallthrough
		case newMsg.Message.Text == "updateDB" && ok:
			log.Printf("UpdateDB started by user: %s\n", newMsg.Message.From.UserName)
			lastmsg = "updateDB"
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Укажи адреса .json файлов через пробел. "+
				"Имена файлов должны соответствовать\nfirst.json\nsecond-part-1.json\nsecond-part-2.json\nthird.json:")
			bot.Send(msg)
		case newMsg.Message.Text != "" && lastmsg == "updateDB" && strings.Contains(newMsg.Message.Text, ".json") && ok:
			lastmsg = ""
			urls := strings.Fields(newMsg.Message.Text)
			for _, url := range urls {
				err := urlLoader(url)
				if err != nil {
					msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, fmt.Sprintf("Невозможно обновить базу с этого URL (%s). Введи опять \"updateDB\".", url))
					bot.Send(msg)
					log.Printf("Wrong URL provided: %s\n", newMsg.Message.Text)
				} else {
					msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Файлы .json обновлены!")
					bot.Send(msg)
					readWordsFromFiles()
					log.Println("Update finished")
				}
			}
		case newMsg.Message.Text == "updateDB" && !ok:
			log.Printf("UpdateDB FAILED by user: %s\n", newMsg.Message.From.UserName)
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Обновлять базу могут только админы!")
			bot.Send(msg)
		case newMsg.Message.Text == "addadmin" && ok:
			lastmsg = "addadmin"
			admins[newMsg.Message.From.UserName] = true
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Введи @username для добавления в админы:")
			bot.Send(msg)
			log.Printf("Add admin initialized by user @%s", newMsg.Message.From.UserName)
		case newMsg.Message.Text != "" && lastmsg == "addadmin" && ok:
			lastmsg = ""
			newadmin := strings.TrimPrefix(newMsg.Message.Text, "@")
			admins[newadmin] = true
			log.Printf("New admin added with username @%s\n", newadmin)
			s := ""
			if strings.HasPrefix(newMsg.Message.Text, "@") {
				s = fmt.Sprintf("Админ %s добавлен!", newMsg.Message.Text)
			} else {
				s = fmt.Sprintf("Админ @%s добавлен!", newMsg.Message.Text)
			}
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, s)
			bot.Send(msg)
		case newMsg.Message.Text == "deladmin" && ok:
			lastmsg = "deladmin"
			admins[newMsg.Message.From.UserName] = true
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Введи @username для удаления из админов:")
			bot.Send(msg)
			log.Printf("Delete admin initialized by user @%s\n", newMsg.Message.From.UserName)
		case newMsg.Message.Text != "" && lastmsg == "deladmin" && ok:
			if strings.ToLower(strings.TrimPrefix(newMsg.Message.Text, "@")) == superadmin {
				lastmsg = ""
				msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, "Нельзя удалить суперадмина!")
				bot.Send(msg)
				log.Printf("User @%s tried to delete @%s", newMsg.Message.From.UserName, superadmin)
			} else {
				lastmsg = ""
				if _, ok := admins[strings.ToLower(strings.TrimPrefix(newMsg.Message.Text, "@"))]; ok {
					delete(admins, strings.ToLower(strings.TrimPrefix(newMsg.Message.Text, "@")))
					msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, fmt.Sprintf("Админ @%s больше не админ", strings.TrimPrefix(newMsg.Message.Text, "@")))
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, fmt.Sprintf("@%s итак не было в списке админов", strings.TrimPrefix(newMsg.Message.Text, "@")))
					bot.Send(msg)
				}
			}
		case strings.ToLower(newMsg.Message.Text) == "admins" && ok:
			adm := ""
			for a := range admins {
				name := fmt.Sprintf("@%s", a)
				adm += name + "; "
			}
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, adm)
			bot.Send(msg)
		case newMsg.Message.Text == "/help" || newMsg.Message.Text == "help":
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, fmt.Sprintf("Короч:\nupdateDB - обновить базу слов\nadmins - список админов\naddadmin - добавить админа\ndeladmin - удалить админа\nВсе команды кроме help работают только от админа АХАХАХАХАХАХАХА"))
			bot.Send(msg)
		default:
			msg := tgbotapi.NewMessage(newMsg.Message.Chat.ID, stringComposer(files.f1, files.f2, files.f3, files.f4))
			bot.Send(msg)
		}
	}
}

func readWordsFromFiles() {
	log.Println("Start reading files...")
	execPath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	filePath := execPath + "/../data/"
	files.f1 = openFile(filePath + file1)
	files.f2 = openFile(filePath + file2)
	files.f3 = openFile(filePath + file3)
	files.f4 = openFile(filePath + file4)
	log.Println("Finished reading files.")
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
		if _, ok := val.(string); ok {
			result = append(result, val.(string))
		}
	}
	file.Close()
	return result
}

func stringComposer(f1, f2, f3, f4 []string) string {
	return fmt.Sprintf("%v %v%v %v", strings.Title(f1[rand.Intn(len(f1))]), strings.Title(f2[rand.Intn(len(f2))]), f3[rand.Intn(len(f3))], strings.Title(f4[rand.Intn(len(f4))]))
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("OK"))
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func urlLoader(url string) error {
	execPath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	filePath := execPath + "/../data/" + path.Base(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Can't load file from %s\n", url)
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if path.Base(url) == file1 || path.Base(url) == file2 || path.Base(url) == file3 || path.Base(url) == file4 {
		out, err := os.Create(filePath)
		if err != nil {
			log.Printf("File not created: %s\n", path.Base(url))
			return err
		} else {
			log.Printf("File %s created\n", path.Base(url))
		}
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Printf("File not copied: %s\n", url)
		} else {
			log.Printf("File %s copied\n", path.Base(url))
		}
		out.Close()
	} else {
		log.Printf("File %s doesn't match schema", path.Base(url))
		err := errors.New("Wrong file name")
		return err
	}
	return err
}
