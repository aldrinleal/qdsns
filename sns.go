package sns

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
)

type Notification struct {
	Message          string `json:"Message"`
	MessageId        string `json:"MessageId"`
	Signature        string `json:"Signature"`
	SignatureVersion string `json:"SignatureVersion"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL"`
	Subject          string `json:"Subject"`
	Timestamp        string `json:"Timestamp"`
	Token            string `json:"Token"`
	TopicArn         string `json:"TopicArn"`
	Type             string `json:"Type"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

type ConfirmSubscriptionResponse struct {
	XMLName         xml.Name `xml:"ConfirmSubscriptionResponse"`
	SubscriptionArn string   `xml:"ConfirmSubscriptionResult>SubscriptionArn"`
	RequestId       string   `xml:"ResponseMetadata>RequestId"`
}

type UnsubscribeResponse struct {
	XMLName   xml.Name `xml:"UnsubscribeResponse"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

func (n *Notification) BuildSignature() []byte {
	var builtSignature bytes.Buffer
	signableKeys := []string{"Message", "MessageId", "Subject", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"}
	for _, key := range signableKeys {
		reflectedStruct := reflect.ValueOf(n)
		field := reflect.Indirect(reflectedStruct).FieldByName(key)
		value := field.String()
		if field.IsValid() && value != "" {
			builtSignature.WriteString(key + "\n")
			builtSignature.WriteString(value + "\n")
		}
	}
	return builtSignature.Bytes()
}

func (n *Notification) VerifySignature() error {
	payloadSignature, err := base64.StdEncoding.DecodeString(n.Signature)
	if err != nil {
		return err
	}

	resp, err := http.Get(n.SigningCertURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	decodedPem, _ := pem.Decode(body)
	if decodedPem == nil {
		return errors.New("The decoded PEM file was empty!")
	}

	parsedCertificate, err := x509.ParseCertificate(decodedPem.Bytes)
	if err != nil {
		return err
	}

	return parsedCertificate.CheckSignature(x509.SHA1WithRSA, n.BuildSignature(), payloadSignature)
}

func (n *Notification) Subscribe() (ConfirmSubscriptionResponse, error) {
	var response ConfirmSubscriptionResponse

	if n.SubscribeURL == "" {
		return response, errors.New("Payload does not have a SubscribeURL!")
	}

	resp, err := http.Get(n.SubscribeURL)

	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)

	if xmlErr != nil {
		return response, xmlErr
	}

	return response, nil
}

func (n *Notification) Unsubscribe() (UnsubscribeResponse, error) {
	var response UnsubscribeResponse
	resp, err := http.Get(n.UnsubscribeURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)
	if xmlErr != nil {
		return response, xmlErr
	}
	return response, nil
}
