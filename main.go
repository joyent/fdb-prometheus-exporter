package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/PierreZ/fdb-prometheus-exporter/models"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var db fdb.Database

func main() {

	apiVersion, err := strconv.Atoi(getEnv("FDB_API_VERSION", "620"))
	if err != nil {
		log.Fatal("cannot parse FDB_API_VERSION from env")
	}
	// Different API versions may expose different runtime behaviors.
	fdb.MustAPIVersion(apiVersion)

	clusterFile := getEnv("FDB_CLUSTER_FILE", "/var/fdb/data/fdb.cluster")

	if _, exists := os.LookupEnv("FDB_CREATE_CLUSTER_FILE"); exists {
		execBash()
	}

	fmt.Println("opening cluster file at", clusterFile)
	dat, err := ioutil.ReadFile(clusterFile)
	if err != nil {
		log.Fatalf("cannot read cluster file")
	}
	fmt.Println(string(dat))

	// Open the default database from the system cluster
	db = fdb.MustOpenDatabase(clusterFile)

	exportWorkload, err := strconv.ParseBool(getEnv("FDB_EXPORT_WORKLOAD", "true"))
	if err != nil {
		log.Fatal("cannot parse FDB_EXPORT_WORLOAD from env")
	}

	listenTo := getEnv("FDB_METRICS_LISTEN", ":8080")
	refreshEvery, err := strconv.Atoi(getEnv("FDB_METRICS_EVERY", "10"))
	if err != nil {
		log.Fatal("cannot parse FDB_METRICS_EVERY from env")
	}

	ticker := time.NewTicker(time.Duration(refreshEvery) * time.Second)
	go func() {
		for range ticker.C {
			//Call the periodic function here.
			models, err := retrieveMetrics()
			if err != nil {
				fmt.Errorf("cannot retrieve metrics from FDB: (%v)", err)
				continue
			}

			fmt.Println("retrieved data")

			if exportWorkload {
				models.ExportWorkload()
			}
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenTo, nil))
}

func retrieveMetrics() (*models.FDBStatus, error) {
	fmt.Println("refreshing metrics")
	rawStatus, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		keyCode := append([]byte{255, 255}, []byte("/status/json")...)
		var k fdb.Key
		k = keyCode
		resp, err := tr.Get(k).Get()
		if err != nil {
			return nil, errors.Wrap(err, "cannot retrieve key")
		}
		if len(resp) == 0 {
			return nil, errors.Wrap(err, "no key for models")
		}

		fdbStatus := models.FDBStatus{}
		err = json.Unmarshal(resp, &fdbStatus)
		if err != nil {
			return nil, errors.Wrap(err, "cannot unmarshal key")
		}
		return fdbStatus, nil
	})

	if err != nil {
		return nil, err
	}

	fdbStatus := rawStatus.(models.FDBStatus)
	return &fdbStatus, nil
}

// getEnv is wrapping os.getenv with a fallback
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func execBash() {
	cmd := exec.Command("/create_cluster_file.bash")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Running command 'create_cluster_file' and waiting for it to finish...\n")
	err = cmd.Run()
	if err != nil {
		slurp, _ := ioutil.ReadAll(stderr)
		fmt.Printf("%s\n", slurp)
		log.Fatalf("cannot run create_cluster_file: %v", err)
	}
	fmt.Println("Command finished")
}
