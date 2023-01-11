package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	port = 8001

	yep        = "yep"
	nope       = "nope"
	gone       = "it a gone"
	badRequest = "idk"

	months = [12]string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}

	daySuffix = regexp.MustCompile("(st|nd|rd|th)$")
)

func IntOrNil(val *int) string {
	if val != nil {
		return strconv.Itoa(*val)
	}
	return "(nil)"
}

// zero indexed
func getMonth(month string) int {
	month = strings.ToLower(month)
	if strings.HasSuffix(month, "yet") {
		month = strings.TrimSuffix(month, "yet")
	}

	for i := 0; i < len(months); i++ {
		if month == months[i] {
			return i
		}
	}

	return -1
}

func getText(parts []string) (*string, error) {
	if !(len(parts) >= 2 && len(parts) <= 4) {
		return nil, errors.New("parts length incorrect")
	}

	if parts[len(parts)-1] != "lol" {
		return nil, errors.New("tld incorrect")
	}

	var month = getMonth(parts[len(parts)-2])
	var day *int = nil

	if len(parts) == 3 {
		dayString := daySuffix.ReplaceAllString(parts[0], "")
		dayTyped, err := strconv.Atoi(dayString)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("unknown day value orig=%s, trimmed=%s", parts[0], dayString))
		}
		if dayTyped <= 0 {
			return nil, errors.New(fmt.Sprintf("day value <0, dateTyped=%d", dayTyped))
		}
		day = &dayTyped
	}

	if len(parts) == 4 {
		if !(parts[0] == "is" && parts[1] == "it") {
			return nil, errors.New(fmt.Sprintf("unknown parts, parts[0]=%s, parts[1]=%s", parts[0], parts[1]))
		}
	}

	now := time.Now()
	log.Printf("req.month=%d, req.date=%s, now.month=%d, now.day=%d", month, IntOrNil(day), now.Month(), now.Day())

	monthDelta := time.Month(month+1) - now.Month()
	if monthDelta == 0 {
		if day == nil {
			return &yep, nil
		} else {
			dayDelta := *day - now.Day()
			if dayDelta == 0 {
				return &yep, nil
			} else if dayDelta > 0 {
				return &nope, nil
			} else if dayDelta < 0 {
				return &gone, nil
			}
		}
	} else if monthDelta > 0 {
		return &nope, nil
	} else if monthDelta < 0 {
		return &gone, nil
	}

	return nil, errors.New(fmt.Sprintf("unhandled request, month=%d, day=%s", month, IntOrNil(day)))
}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request, host=%s", r.Host)

	url := r.Host
	parts := strings.Split(url, ".")
	text, err := getText(parts)

	if err != nil {
		log.Printf("failed to parse request %v", err)
		w.WriteHeader(http.StatusBadRequest)
		text = &badRequest
	}

	_, err = io.WriteString(w, *text)
	if err != nil {
		log.Println("failed to write response", err)
	}
}

func main() {
	log.Printf("starting server at port=%d\n", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)
	server := &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(port), Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("http.ListenAndServe failed", err)
	}
}
