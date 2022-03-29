package models

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	databaseClientCount = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "fdb_client_count",
                Help: "number of connected clients",
        })

	databaseStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fdb_database_status",
		Help: "state of the dabase",
	}, []string{"state"})
)

// ExportDatabaseStatus is exporting the status of the database
func (s FDBStatus) ExportDatabaseStatus() {

	databaseClientCount.Set(float64(s.Cluster.Clients.Count))

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
	r.MustRegister(databaseStatus)
	r.MustRegister(databaseClientCount)
}
