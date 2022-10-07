package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func replyJSON(w http.ResponseWriter, content any, logger *log.Logger) error {
	bytes, err := marshalJSON(content)
	if err != nil {
		logger.Panicf("Cannot marshal response: %s", err.Error())
		return err
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Encoding, Authorization, Content-Type, Content-Length, Origin, X-Requested-With, X-CSRF-Token")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(bytes)
	logger.Printf("Bytes written in reply json: %s", string(bytes))
	if err != nil {
		logger.Panicf("Cannot write bytes in replyJSON: %s", err.Error())
		return err
	}
	return nil
}

func marshalJSON(content any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(content)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalJSON(bytes []byte) (any, error) {
	var ans any
	err := json.Unmarshal(bytes, &ans)
	return ans, err
}

func postJSON(addr string, content any, logger *log.Logger) (*http.Response, error) {
	thebytes, err := marshalJSON(content)
	if err != nil {
		logger.Printf("Cannot marshal post JSON: %s", err.Error())
		return nil, err
	}
	r := bytes.NewReader(thebytes)
	fullAddr := "http://" + addr
	logger.Printf("Posting %s to %s", thebytes, fullAddr)

	req, err := http.NewRequest("POST", fullAddr, r)
	if err != nil {
		logger.Printf("Cannot build post request : %s", err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		logger.Printf("Cannot post JSON to %s : %s", addr, err.Error())
		return nil, err
	}
	return resp, nil
}

func decodeJSONResponse(resp *http.Response, logger *log.Logger) (any, error) {

	bodyParsed, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.Printf("Cannot decode JSON ans: %s", err.Error())
		return nil, err
	}

	ans, err := unmarshalJSON(bodyParsed)
	if err != nil {
		logger.Printf("Cannot unmarshal JSON ans : %s", err.Error())
		return nil, err
	}
	return ans, nil
}

func decodeIntResponse(resp *http.Response, logger *log.Logger) (int, error) {
	bodyParsed, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.Printf("Cannot decode int ans: %s", err.Error())
		return 0, err
	}

	ans := string(bodyParsed)
	integer, err := strconv.Atoi(ans)
	if len(ans) == 0 || err != nil {
		logger.Printf("Cannot unmarshal int ans : %s", err.Error())
		return 0, err
	}

	return integer, nil
}
