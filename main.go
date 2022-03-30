package main

import (
	"apilimit/debuglog"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type WafPolicy struct {
	APIRateLimit []*APIRatelimit
}

var log *logrus.Logger = &logrus.Logger{}

type apiratelimitupdate struct {
	Orgid    int `json:"orgid"`
	Domainid int `josn:"domainid"`
}
type APIRatelimit struct {
	Enable          bool            `json:"enable"`
	ID              int64           `json:"id"`
	Name            string          `json:"name"`
	DetectClientIP  DetectClientIP  `json:"detect_client_ip"`
	DetectTarget    DetectTarget    `json:"detect_target"`
	DetectCondition DetectCondition `json:"detect_condition"`
	DetectBehavior  DetectBehavior  `json:"detect_behavior"`
	Explain         string          `json:"explain"`
	Limiter         *rate.Limiter
}

type DetectBehavior struct {
	Action   string `json:"action"`
	Log      bool   `json:"log"`
	Mail     bool   `json:"mail"`
	Severity string `json:"severity"`
	Page     int64  `json:"page"`
}

type DetectClientIP struct {
	ApplyIP       []string `json:"apply_ip"`
	ExceptIP      []string `json:"except_ip"`
	ApplyIPGroup  []int64  `json:"apply_ip_group"`
	ExceptIPGroup []int64  `json:"except_ip_group"`
}

type DetectCondition struct {
	Totalcount int64    `json:"totalcount"`
	Method     []string `json:"method"`
}

type DetectTarget struct {
	Path []string `json:"path"`
}

// func ApiRateLimitUpdate(rw http.ResponseWriter, req *http.Request) {
func ApiRateLimitUpdate() {
	// if req.Method != "POST" {
	// 	logrus.Info("Not Supported")
	// 	return
	// }

	// body, err := ioutil.ReadAll(io.LimitReader(req.Body, 10*1024*1024))
	// if err != nil {
	// 	logrus.Panic(err)
	// }
	// log.Infof("body : %s", body)

	// var info apiratelimitupdate

	// err = json.Unmarshal(body, &info)
	// if err != nil {
	// 	log.Errorf("Unmarshal Error at WebcacheUpdate : %v", err)
	// 	return
	// }
	currpath, pwderr := os.Getwd()
	if pwderr != nil {
		log.Errorf("Cannot get Current Directory : %s", pwderr)
	}
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/api_ratelimit.json", currpath))
	if err != nil {
		log.Errorf("Cannot read file : %s", err)
		return
	}
	if policy := ApiRateLimitParser(data); policy != nil {
		log.Infof("Enable : %v, ID : %v, Name : %v", policy.Enable, policy.ID, policy.Name)
		log.Infof("DetectClientIP.ApplyIP : %v, DetectClient.ApplyIPGroup", policy.DetectClientIP.ApplyIP, policy.DetectClientIP.ApplyIPGroup)
		log.Infof("DetectClientIP.ExceptIP : %v, DetectClient.ExceptIPGroup", policy.DetectClientIP.ExceptIP, policy.DetectClientIP.ExceptIPGroup)
		log.Infof("DetectTarget : %v", policy.DetectTarget)
		log.Infof("DetectCondition.Totalcount : %v, DetectCondition.Method : %v", policy.DetectCondition.Totalcount, policy.DetectCondition.Method)
		log.Infof("DetectBehavior.Action : %v, DetectBehavior.Log : %v, policy.DetectBehavior.Mail : %v, policy.DetectBehavior.Severity : %v, policy.DetectBehavior.Page : %v", policy.DetectBehavior.Action, policy.DetectBehavior.Log, policy.DetectBehavior.Mail, policy.DetectBehavior.Severity, policy.DetectBehavior.Page)
		log.Infof("Explain : %v", policy.Explain)
		log.Info("SUCCESS TO GET POLICY INFO")
		ApiRateLimitpolicy = policy

	}
}

var ApiRateLimitpolicy *APIRatelimit

func ApiRateLimitParser(body []byte) *APIRatelimit {
	var temppolicy *APIRatelimit

	err := json.Unmarshal(body, &temppolicy)
	if err != nil {
		log.Errorf("Cannot get policy : %s", err)
		return nil
	}

	temppolicy.Limiter = rate.NewLimiter(rate.Every(1*time.Second), int(temppolicy.DetectCondition.Totalcount))

	return temppolicy
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", okHandler)

	log.Info("Listening on :4000...")
	http.ListenAndServe(":4000", limit(mux))
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("OK")
	w.Write([]byte("OK\n"))
}

// Removing old entries from the map

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limimter *rate.Limiter
	lastSeen time.Time
}

// Change the the map to hold values of the type visitor
var visitors = make(map[string]*visitor)
var mu sync.Mutex

func init() {

	log = debuglog.DebugLogInit("aplimit")
	setLevel("INFO")
	ApiRateLimitUpdate()
}

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if ApiRateLimitpolicy.Limiter.Allow() == false {
			log.Info("TOO MANY REQUESTS!!!")
			http.Error(w, "TOO MANY REQUESTS!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setLevel(lvl string) {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		level = logrus.ErrorLevel
	}
	fmt.Fprint(os.Stderr, "set log level\n", level.String())
	log.SetLevel(level)
}
