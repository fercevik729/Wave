/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"bytes"
	"reflect"
	"testing"
)

func TestWhirlpool(t *testing.T) {
	requests := make([]Request, 0)
	requests = append(requests, Request{
		reqType:  "GET",
		endpoint: "https://api.sampleapis.com/coffee/hot",
		body:     bytes.Buffer{},
		isAuth:   false,
	}, Request{
		reqType:  "POST",
		endpoint: "https://postman-echo.com/post",
		body:     *readJsonFile("../data/post.json"),
		isAuth:   false,
	})
	actual := Whirlpool(10, requests, false, "", KeyChain{})
	expected := 20
	if actual != expected {
		t.Errorf("Expected %d successes, but got %d successes\n", actual, expected)
	}
}

func TestSplash(t *testing.T) {
	requests := make([]Request, 0)
	requests = append(requests, Request{
		reqType:  "GET",
		endpoint: "https://api.sampleapis.com/coffee/hot",
		body:     bytes.Buffer{},
		isAuth:   false,
	}, Request{
		reqType:  "POST",
		endpoint: "https://postman-echo.com/post",
		body:     *readJsonFile("../data/post.json"),
		isAuth:   false,
	})
	actual := Splash(10, requests, true, "", KeyChain{})
	expected := 20
	if actual != expected {
		t.Errorf("Expected %d successes, but got %d successes\n", actual, expected)
	}
}

func TestNew(t *testing.T) {
	actReqs, actChain := New("../requests/test-requests.txt", "../data/cred.yaml")
	expReqs := []Request{
		{
			reqType:  "GET",
			endpoint: "https://api.sampleapis.com/coffee/hot",
			body:     bytes.Buffer{},
			isAuth:   false,
		}, {
			reqType:  "POST",
			endpoint: "https://postman-echo.com/post",
			body:     *readJsonFile("../data/post.json"),
			isAuth:   false,
		}}
	if !reflect.DeepEqual(actReqs, expReqs) {
		t.Errorf("Expected and actual requests are not equal")
	}
	expChain := KeyChain{
		User:  "developer45@gmail.com",
		Pass:  "password1234",
		Token: "Bearer xxxxxxxxxxxxxxxxxxxxxxxx",
	}
	if actChain != expChain {
		t.Errorf("Expected and actual keychains are not equal")
	}
}

func TestNewYAMLRequests(t *testing.T) {
	actualReqs := NewYAMLRequests("../requests/reqs.yaml")
	expectedReqs := make(map[string]YAMLRequest, 0)

	expectedReqs["request-1"] = YAMLRequest{
		Method:      "GET",
		Base:        "https://api.sampleapis.com",
		Endpoint:    "/coffee/hot",
		SuccessCode: 200,
	}
	expectedReqs["request-2"] = YAMLRequest{
		Method:      "POST",
		Base:        "https://postman-echo.com",
		Endpoint:    "/post",
		SuccessCode: 200,
		DataFile:    "./data/post.json",
		ContentType: "application/json",
		IsAuth:      true,
		RToken:      false,
	}

	if !reflect.DeepEqual(actualReqs, expectedReqs) {
		t.Errorf("Expected %v, got %v", actualReqs, expectedReqs)
	}
}
