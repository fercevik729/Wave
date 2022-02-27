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
