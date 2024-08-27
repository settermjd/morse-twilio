package main

import (
	"encoding/xml"
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
)

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

type TwiML struct {
	XMLName xml.Name `xml:"Response"`
	Play    []Play   `xml:"Play"`
}

type Play struct {
	Digits string `xml:"Digits,attr"`
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

func morseToTwiML(morse string) TwiML {
	var response TwiML
	for _, char := range morse {
		switch char {
		case '.':
			response.Play = append(response.Play, Play{Digits: "1"})
		case '-':
			response.Play = append(response.Play, Play{Digits: "9"})
		case ' ':
			response.Play = append(response.Play, Play{Digits: "w"})
		}
	}
	return response
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	message := r.FormValue("message")
	phoneNumber := r.FormValue("phone_number")
	encodedMessage := url.QueryEscape(message)
	baseUrl := "https://aee2-87-121-75-154.ngrok-free.app"
	params := &twilioApi.CreateCallParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetUrl(baseUrl + "/voice?message=" + encodedMessage)
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
	morse := textToMorse(message)
	twiml := morseToTwiML(morse)
	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(twiml)
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

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/voice", handleVoiceRequest)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
