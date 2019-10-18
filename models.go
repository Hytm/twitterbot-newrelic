package main

//LanguageDetectionResponse is the default answer by language detection service from Azure
type LanguageDetectionResponse struct {
	Documents []struct {
		DetectedLanguages []struct {
			Iso6391Name string `json:"iso6391Name"`
			Name        string `json:"name"`
			Score       int    `json:"score"`
		} `json:"detectedLanguages"`
		ID string `json:"id"`
	} `json:"documents"`
	Errors []interface{} `json:"errors"`
}

//SentimentDetectionResponse is the default answer by language detection service from Azure
type SentimentDetectionResponse struct {
	Documents []struct {
		ID    string  `json:"id"`
		Score float64 `json:"score"`
	} `json:"documents"`
	Errors []interface{} `json:"errors"`
}
