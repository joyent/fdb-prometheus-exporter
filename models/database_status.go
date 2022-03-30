package models

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	databaseClientCount = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "fdb_client_count",
                Help: "number of connected clients",
        })

	databaseDiskUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fdb_database_data_size_bytes",
		Help: "number of data bytes used",
	}, []string{"usage_type"})

	databaseLatencyProbe = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fdb_latency_probe",
		Help: "latency values based on running sample transactions",
	}, []string{"probe"})

	databasePartitionCount = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "fdb_partition_count",
                Help: "number of fdb partitions",
        })

	databaseStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fdb_database_status",
		Help: "state of the dabase",
	}, []string{"state"})
)

// ExportDatabaseStatus is exporting the status of the database
func (s FDBStatus) ExportDatabaseStatus() {

	databaseClientCount.Set(float64(s.Cluster.Clients.Count))

	databaseDiskUsed.With(prometheus.Labels{
		"usage_type": "totalDisk",
	}).Set(float64(s.Cluster.Data.TotalDiskUsedBytes))
	databaseDiskUsed.With(prometheus.Labels{
		"usage_type": "totalKv",
	}).Set(float64(s.Cluster.Data.TotalKvSizeBytes))
	databaseDiskUsed.With(prometheus.Labels{
		"usage_type": "systemKv",
	}).Set(float64(s.Cluster.Data.SystemKvSizeBytes))

	databaseLatencyProbe.With(prometheus.Labels{
		"probe": "batch_priority_transaction_start_seconds",
	}).Set(float64(s.Cluster.LatencyProbe.BatchPriorityTransactionStartSeconds))
	databaseLatencyProbe.With(prometheus.Labels{
		"probe": "commit_seconds",
	}).Set(float64(s.Cluster.LatencyProbe.CommitSeconds))
	databaseLatencyProbe.With(prometheus.Labels{
		"probe": "immediate_priority_transaction_start_seconds",
	}).Set(float64(s.Cluster.LatencyProbe.ImmediatePriorityTransactionStartSeconds))
	databaseLatencyProbe.With(prometheus.Labels{
		"probe": "read_seconds",
	}).Set(float64(s.Cluster.LatencyProbe.ReadSeconds))
	databaseLatencyProbe.With(prometheus.Labels{
		"probe": "transaction_start_seconds",
	}).Set(float64(s.Cluster.LatencyProbe.TransactionStartSeconds))

	databasePartitionCount.Set(float64(s.Cluster.Data.PartitionsCount))

	databaseStatus.With(prometheus.Labels{
		"state": "available",
	}).Set(boolToNumber(s.Client.DatabaseStatus.Available))
	databaseStatus.With(prometheus.Labels{
		"state": "healthy",
	}).Set(boolToNumber(s.Client.DatabaseStatus.Healthy))
	databaseStatus.With(prometheus.Labels{
		"state": "quorum_reachable",
	}).Set(boolToNumber(s.Client.Coordinators.QuorumReachable))
	databaseStatus.With(prometheus.Labels{
		"state": "locked",
	}).Set(boolToNumber(s.Cluster.DatabaseLocked))
}

func registerDatabaseStatus(r *prometheus.Registry) {
	r.MustRegister(databaseClientCount)
	r.MustRegister(databaseDiskUsed)
	r.MustRegister(databaseLatencyProbe)
	r.MustRegister(databasePartitionCount)
	r.MustRegister(databaseStatus)
}
