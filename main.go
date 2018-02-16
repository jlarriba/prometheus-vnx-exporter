package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	yaml "gopkg.in/yaml.v2"
)

var (
	addr       = flag.String("listen-address", ":9184", "The address to listen on for HTTP requests.")
	storageMng = flag.String("storage", "10.10.10.10", "The address of the vnx manager")
	user       = flag.String("user", "user", "The user used to access the vnx manager")
	password   = flag.String("password", "password", "The password used to access the vnx manager")
	poolname   = flag.String("poolname", "pool", "The pool name of the vnx manager")
)

var (
	availableCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_available_capacity_in_gb",
		Help: "Available Capacity (GBs)",
	})
	consumedCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_consumed_capacity_in_gb",
		Help: "Consumed Capacity (GBs)",
	})
	userCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_user_capacity_in_gb",
		Help: "User Capacity (GBs)",
	})
	percentFull = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_percent_full",
		Help: "Percent Full",
	})
	totalSubscribed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_total_subscribed_capacity_in_gb",
		Help: "Total Subscribed Capacity (GBs)",
	})
	percentSubscribed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_percent_subscribed",
		Help: "Percent Subscribed",
	})
	consumedNumOfLun = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_consumed_num_of_lun",
		Help: "Consumed Number of LUN",
	})
	maximumNumOfLun = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vnx_maximum_num_of_lun",
		Help: "Maximum Number of LUN",
	})
)

type storageData struct {
	AvailableCapacity float64 `yaml:"Available Capacity (GBs)"`
	ConsumedCapacity  float64 `yaml:"Consumed Capacity (GBs)"`
	UserCapacity      float64 `yaml:"User Capacity (GBs)"`
	PercentFull       float64 `yaml:"Percent Full"`
	TotalSubscribed   float64 `yaml:"Total Subscribed Capacity (GBs)"`
	PercentSubscribed float64 `yaml:"Percent Subscribed"`
}

func init() {
	prometheus.MustRegister(availableCapacity)
	prometheus.MustRegister(consumedCapacity)
	prometheus.MustRegister(userCapacity)
	prometheus.MustRegister(percentFull)
	prometheus.MustRegister(totalSubscribed)
	prometheus.MustRegister(percentSubscribed)
	prometheus.MustRegister(consumedNumOfLun)
	prometheus.MustRegister(maximumNumOfLun)
}

func main() {
	flag.Parse()
	maximumNumOfLun.Set(1100)
	go getStorageMetrics()
	go getLunMetrics()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func getStorageMetrics() {
	for {
		cmd := exec.Command("bash", "-c", "/opt/Navisphere/bin/naviseccli -h "+*storageMng+" -user "+*user+" -password "+*password+" -Scope 0 storagepool  -list -name "+*poolname+" |head -26")
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		data := new(storageData)
		err2 := yaml.Unmarshal(out, data)
		if err2 != nil {
			log.Fatalf("error: %v", err2)
		}
		availableCapacity.Set(data.AvailableCapacity)
		consumedCapacity.Set(data.ConsumedCapacity)
		userCapacity.Set(data.UserCapacity)
		percentFull.Set(data.PercentFull)
		totalSubscribed.Set(data.TotalSubscribed)
		percentSubscribed.Set(data.PercentSubscribed)

		time.Sleep(60 * time.Second)
	}
}

func getLunMetrics() {
	for {
		cmd := exec.Command("bash", "-c", "/opt/Navisphere/bin/naviseccli -h "+*storageMng+" -user "+*user+" -password "+*password+" -Scope 0  getlun |grep LOGICAL|wc -l")
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		fout, err2 := strconv.ParseFloat(string(out[:len(out)-1]), 64)
		if err2 != nil {
			log.Fatal(err2)
		}
		consumedNumOfLun.Set(fout)

		time.Sleep(60 * time.Second)
	}
}
