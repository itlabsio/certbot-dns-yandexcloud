package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"crypto/rsa"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"strings"
)

var KEY_FILE string = os.Getenv("KEY_FILE")
var KEY_ID string = os.Getenv("KEY_ID")
var SERVICE_ACCOUNT_ID string = os.Getenv("SERVICE_ACCOUNT_ID")
var CLOUD_ID string = os.Getenv("CLOUD_ID")
var API_MAIN_DOMAIN string = "api.cloud.yandex.net/"
var DNS_ZONE_ID string = os.Getenv("DNS_ZONE_ID")
var DNS_RECORD_NAME string = os.Getenv("DNS_RECORD_NAME")
var DNS_RECORD_TYPE string = getEnv("DNS_RECORD_TYPE", "TXT")
var DNS_RECORD_TTL string = getEnv("DNS_RECORD_TTL", "60")
var DNS_RECORD_DATA string = os.Getenv("CERTBOT_VALIDATION")

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func checkEnvVars() error {
	if KEY_ID == "" {
		return errors.New("variable KEY_ID not set or empty")
	}

	if SERVICE_ACCOUNT_ID == "" {
		return errors.New("variable SERVICE_ACCOUNT_ID not set or empty")
	}

	if CLOUD_ID == "" {
		return errors.New("variable CLOUD_ID not set or empty")
	}

	if DNS_RECORD_DATA == "" {
		return errors.New("variable CERTBOT_VALIDATION not set or empty")
	}

	return nil
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func makeHttpRequest(requestType, requestUrl string, requestPayload io.Reader, iamToken string) []byte {
	client := &http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, requestType, requestUrl, requestPayload)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+iamToken)

	resp, err := client.Do(req)
	checkError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)

	return body
}

func signedToken() string {
	claims := jwt.RegisteredClaims{
		Issuer:    SERVICE_ACCOUNT_ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	token.Header["kid"] = KEY_ID

	privateKey := loadPrivateKey()
	signed, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	return signed
}

func loadPrivateKey() *rsa.PrivateKey {
	data, err := ioutil.ReadFile(KEY_FILE)
	if err != nil {
		panic(err)
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		panic(err)
	}
	return rsaPrivateKey
}

func getIAMToken() string {
	jot := signedToken()
	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jot)),
	)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("%s: %s", resp.Status, body))
	}

	var data struct {
		IAMToken string `json:"iamToken"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	return data.IAMToken
}

func createDnsRecordInYandex(zoneId string, dnsRecordName string, dnsRecodType string, dnsRecordTtl string, dnsRecordData []string, iamToken string) error {
	type Replacements struct {
		Name string   `json:"name"`
		Type string   `json:"type"`
		TTL  string   `json:"ttl"`
		Data []string `json:"data"`
	}

	type Payload struct {
		Replacements []Replacements `json:"replacements"`
	}

	type YandexResponseRecordAction struct {
		Done bool `json:"done"`
	}

	replacements := Replacements{dnsRecordName, dnsRecodType, dnsRecordTtl, dnsRecordData}
	replacementsArray := []Replacements{replacements}
	payload := Payload{replacementsArray}

	payloadBytes, err := json.Marshal(payload)
	checkError(err)

	var yandexResponse = YandexResponseRecordAction{}
	var apiEndpoint string = "https://dns." + API_MAIN_DOMAIN + "dns/v1/zones/"
	var body = makeHttpRequest("POST", apiEndpoint+zoneId+":upsertRecordSets", bytes.NewBuffer(payloadBytes), iamToken)

	err = json.Unmarshal(body, &yandexResponse)
	checkError(err)

	if !yandexResponse.Done {
		log.Println(string(body))
		return errors.New("Record " + dnsRecordName + dnsRecodType + dnsRecordTtl + strings.Join(dnsRecordData, " ") + " in dns zone " + zoneId + " not created. Or bad created")
	}
	return nil
}

func main() {
	// Check env variabels
	var err error = checkEnvVars()
	checkError(err)

	// Configure logger to JSON format
	log.SetFormatter(&log.JSONFormatter{})

	// Get IAM token via service account private key
	iamToken := getIAMToken()

	DNS_RECORD_DATA := []string{"\"" + DNS_RECORD_DATA + "\""}

	// Try create DNS record
	err = createDnsRecordInYandex(DNS_ZONE_ID, DNS_RECORD_NAME, DNS_RECORD_TYPE, DNS_RECORD_TTL, DNS_RECORD_DATA, iamToken)
	checkError(err)

	// Sleep for renewing DNS record for certbot
	time.Sleep(time.Second * 15)

}
