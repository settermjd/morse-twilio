package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
)

const baseUrl = "https://aee2-87-121-75-154.ngrok-free.app"

var client *twilio.RestClient

var MorseCode = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..",
	'E': ".", 'F': "..-.", 'G': "--.", 'H': "....",
	'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..",
	'M': "--", 'N': "-.", 'O': "---", 'P': ".--.",
	'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-",
	'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-",
	'Y': "-.--", 'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--",
	'4': "....-", '5': ".....", '6': "-....", '7': "--...",
	'8': "---..", '9': "----.",
}

func textToMorse(text string) string {
	var result strings.Builder
	for _, char := range strings.ToUpper(text) {
		if code, ok := MorseCode[char]; ok {
			result.WriteString(code)
			result.WriteString(" ")
		} else if char == ' ' {
			result.WriteString("  ")
		}
	}
	return result.String()
}

func morseToTwiML(morse string) []twiml.Element {
	var verbList []twiml.Element
	for _, char := range morse {
		switch char {
		case '.':
			verbList = append(verbList, &twiml.VoicePlay{Digits: "1"})
		case '-':
			verbList = append(verbList, &twiml.VoicePlay{Digits: "9"})
		case ' ':
			verbList = append(verbList, &twiml.VoicePlay{Digits: "w"})
		}
	}
	return verbList
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	message := r.FormValue("message")
	phoneNumber := r.FormValue("phone_number")
	params := &twilioApi.CreateCallParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetUrl(baseUrl + "/voice?message=" + url.QueryEscape(message))

	_, err := client.Api.CreateCall(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Call initiated to %s with message: %s", phoneNumber, message)
}

func handleVoiceRequest(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("message")
	if message == "" {
		message = "Hello World"
	}
	w.Header().Set("Content-Type", "application/xml")

	output, err := twiml.Voice(morseToTwiML(textToMorse(message)))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write([]byte(output))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("POST /submit", handleSubmit)
	mux.HandleFunc("/voice", handleVoiceRequest)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
