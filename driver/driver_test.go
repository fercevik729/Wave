/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"reflect"
	"testing"
)

func TestWhirlpool(t *testing.T) {
	reqs := make(map[string]*Request, 0)
	reqs["request-1"] = &Request{
		Method:      "GET",
		Base:        "https://api.sampleapis.com",
		Endpoint:    "/coffee/hot",
		SuccessCode: 200,
	}
	reqs["request-2"] = &Request{
		Method:      "POST",
		Base:        "https://postman-echo.com",
		Endpoint:    "/post",
		SuccessCode: 200,
		DataFile:    "./data/post.json",
		ContentType: "application/json",
		IsAuth:      false,
		RToken:      false,
	}
	actual := Whirlpool(10, reqs, false, "", KeyChain{})
	expected := 20
	if actual != expected {
		t.Errorf("Expected %d successes, but got %d successes\n", expected, actual)
	}
}

func TestSplash(t *testing.T) {
	reqs := make(map[string]*Request, 0)
	reqs["request-1"] = &Request{
		Method:      "GET",
		Base:        "https://api.sampleapis.com",
		Endpoint:    "/coffee/hot",
		SuccessCode: 200,
	}
	reqs["request-2"] = &Request{
		Method:      "POST",
		Base:        "https://postman-echo.com",
		Endpoint:    "/post",
		SuccessCode: 200,
		DataFile:    "./data/post.json",
		ContentType: "application/json",
		IsAuth:      true,
		RToken:      false,
	}

	actual := Splash(10, reqs, true, "", KeyChain{})
	expected := 20
	if actual != expected {
		t.Errorf("Expected %d successes, but got %d successes\n", actual, expected)
	}
}

func TestNew(t *testing.T) {
	// TODO: Update test
	actualReqs, actChain := New("../requests/test-reqs.yaml", "../data/cred.yaml")
	expectedReqs := make(map[string]*Request, 0)
	expectedChain := KeyChain{
		User:  "developer45@gmail.com",
		Pass:  "password1234",
		Token: "Bearer xxxxxxxxxxxxxxxxxxxxxxxx",
	}

	expectedReqs["request-1"] = &Request{
		Method:      "GET",
		Base:        "https://api.sampleapis.com",
		Endpoint:    "/coffee/hot",
		SuccessCode: 200,
	}
	expectedReqs["request-2"] = &Request{
		Method:      "POST",
		Base:        "https://postman-echo.com",
		Endpoint:    "/post",
		SuccessCode: 200,
		ContentType: "application/json",
	}

	if !reflect.DeepEqual(actualReqs, expectedReqs) {
		t.Errorf("Requests: expected %v, but got %v", expectedReqs, actualReqs)
	}
	if actChain != expectedChain {
		t.Errorf("Keychain: expected %v, but got %v", expectedChain, actChain)
	}
}
