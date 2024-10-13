package main

import (
	"bufio"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	UserInput  string
	Accuracy   int
	WPM        float64
	TextToType string
}

var (
	words     []string
	startTime time.Time
)

func init() {
	words = loadWordsFromFile("words/words.txt")
}

func loadWordsFromFile(filename string) []string {
	var words []string
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening words file:", err)
		return words
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading words file:", err)
	}

	return words
}

func main() {
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/results", resultsHandler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	startTime = time.Now()
	randomWords := getRandomWords(10)
	textToType := strings.Join(randomWords, " ")
	result := Result{TextToType: textToType}
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, result)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userInput := r.FormValue("userInput")
		textToType := r.FormValue("textToType")

		accuracy := calculateAccuracy(userInput, textToType)

		elapsed := time.Since(startTime).Minutes()
		wordCount := len(strings.Fields(userInput))
		wpm := float64(wordCount) / elapsed

		http.Redirect(w, r, fmt.Sprintf("/results?input=%s&accuracy=%d&wpm=%.2f&textToType=%s",
			userInput, accuracy, wpm, textToType), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	userInput := r.URL.Query().Get("input")
	accuracy := r.URL.Query().Get("accuracy")
	wpm := r.URL.Query().Get("wpm")
	textToType := r.URL.Query().Get("textToType")

	accuracyInt, _ := strconv.Atoi(accuracy)
	wpmFloat, _ := strconv.ParseFloat(wpm, 64)

	result := Result{
		UserInput:  userInput,
		Accuracy:   accuracyInt,
		WPM:        wpmFloat,
		TextToType: textToType,
	}

	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, result)
}

func getRandomWords(n int) []string {
	rand.Seed(time.Now().UnixNano())
	selectedWords := rand.Perm(len(words))[:n]
	randomWords := make([]string, n)
	for i, idx := range selectedWords {
		randomWords[i] = words[idx]
	}
	return randomWords
}

func calculateAccuracy(userInput, originalText string) int {
	if len(originalText) == 0 {
		return 0
	}

	correctChars := 0
	incorrectChars := 0

	for i := 0; i < len(originalText); i++ {
		if i < len(userInput) {
			if userInput[i] == originalText[i] {
				correctChars++
			} else {
				incorrectChars++
			}
		} else {
			incorrectChars++
		}
	}

	totalTypedChars := correctChars + incorrectChars + (len(userInput) - len(originalText))
	return int(float64(correctChars) / float64(totalTypedChars) * 100)
}
