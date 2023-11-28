package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendTelegramNotification(bot *tgbotapi.BotAPI, bodyString string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred in sendTelegramNotification:", r)
		}
	}()

	log.Printf("Bot username: %s", bot.Self.UserName)
	log.Printf("Bot ID: %d", bot.Self.ID)

	chatID := int64(112845421) // YOUR_TELEGRAM_CHAT_ID
	//mio = 112845421
	//gruppo = -974313836
	//supergruppo = -1001946027674

	result := GetCharactersAfterSubstring(bodyString, "data=")
	fmt.Println("Characters after the substring:", result)

	today := time.Now().In(time.FixedZone("EST", -5*3600))

	// Convert the target date to a time.Time object
	targetDate, err := time.Parse("02-01-2006", result)
	if err != nil {
		panic(err)
	}

	// Calculate the number of days between the two dates
	daysUntilTargetDate := targetDate.Sub(today).Hours() / 24
	fmt.Println("daysUntilTargetDate = ", daysUntilTargetDate)

	if daysUntilTargetDate > 155 {
		LogToWebSocket("Ancora niente - prossimo check fra 8 minuti")
	} else {
		LogToWebSocket("TROVATO UN POSTO - INVIO MESSAGGIO SU TELEGRAM")

		// Create the Telegram message without the file
		msg := tgbotapi.NewMessage(chatID, "Trovato un posto"+"    \n\ndata = "+result)

		// Send the message
		_, err = bot.Send(msg)
		if err != nil {
			log.Println("Error sending Telegram message:", err)
		}
	}
}

func GetCharactersAfterSubstring(inputString, substring string) string {
	index := strings.Index(inputString, substring)

	if index == -1 {
		// Substring not found, return an empty string or handle the error accordingly.
		return ""
	}

	endPosition := index + len(substring) + 10

	// Check if the end position is within the bounds of the inputString.
	if endPosition > len(inputString) {
		endPosition = len(inputString)
	}

	// Extract the characters after the substring up to the 10th character.
	return inputString[index+len(substring) : endPosition]
}

func HandleTriggerRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Received trigger request - polling API and sending Telegram notification")

	// Extract the combined cookies from the URL parameters
	cookies :=
		"JSESSIONID=" + r.URL.Query().Get("JSESSIONID")

	cookies = strings.ReplaceAll(cookies, "%3B", ";")
	cookies = strings.ReplaceAll(cookies, "%26", "&")
	cookies = strings.ReplaceAll(cookies, " ", "")

	log.Println("cookies = " + cookies)

	if !CheckAPI(cookies) {
		SendErrorResponse(w, "Error: Invalid or expired cookies. Please try again with valid cookies.")
		return
	}

	bot, err := tgbotapi.NewBotAPI("5878994522:AAGAgNPCncWJxgMou5q0x6UOgkyUuD_99VA")
	if err != nil {
		log.Println("Error initializing Telegram bot:", err)
		SendErrorResponse(w, "Error initializing Telegram bot. Please check the provided API token.")
		return
	}

	pollAPIFlag = true
	go PollAPI(w, bot, cookies) // Start the API polling in a separate goroutine

	// Respond to the trigger request with a success message
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("API trigger received. Polling has started with the provided cookies."))
}

func KeepAlive() {
	for {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://passaporto.onrender.com/", nil)
		if err != nil {
			log.Println("Error creating request to Render instance:", err)
			time.Sleep(pollingTime) // Retry after the pollingTime
			continue
		}

		// Make the API call to the Render instance
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error calling Render instance:", err)
		} else {
			defer res.Body.Close()
			log.Println("API call to Render instance successful")
		}

		time.Sleep(pollingTime)
	}
}