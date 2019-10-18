package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

func makeCall(uri string, data []map[string]string) (body []byte, err error) {
	d, err := json.Marshal(&data)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		return body, err
	}

	r := strings.NewReader("{\"documents\": " + string(d) + "}")

	c := &http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest("POST", uri, r)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return body, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", azTextAnalyticsKey)

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error on request: %v\n", err)
		return body, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v\n", err)
		return body, err
	}

	return body, err
}
func detectLanguage(uri, message string, txn newrelic.Transaction) string {
	defer newrelic.StartSegment(txn, "detectLanguage").End()
	data := []map[string]string{
		{"id": "1", "text": message},
	}

	body, err := makeCall(uri, data)
	if err != nil {
		return ""
	}

	var f LanguageDetectionResponse
	json.Unmarshal(body, &f)

	l := ""
	if len(f.Documents) > 0 {
		l = f.Documents[0].DetectedLanguages[0].Iso6391Name
	}
	return l
}

func detectSentiment(uri, lang, message string, txn newrelic.Transaction) (bool, float64) {
	defer newrelic.StartSegment(txn, "detectSentiment").End()
	data := []map[string]string{
		{"id": "1", "language": lang, "text": message},
	}

	body, err := makeCall(uri, data)
	if err != nil {
		return false, 0.0
	}

	var f SentimentDetectionResponse
	json.Unmarshal(body, &f)

	//If the score is close to 1 this is a positive. If the score is less than threshold , we declare this is a negative feedback
	if len(f.Documents) > 0 {
		if f.Documents[0].Score < threshold {
			return true, f.Documents[0].Score
		}
	}
	return false, 0.0
}
