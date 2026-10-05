package metrics

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The database collector (ADR 0123 §3, "Game engine health"):
//
//	cmdctrl_db_size_bytes                              the live SQLite file
//	cmdctrl_db_backup_last_success_timestamp_seconds   the last VACUUM INTO backup
//
// main.go registers it only when there is a database; with none, both
// are absent.

// DBSource is the database (*db.DB).
type DBSource interface {
	// SizeBytes is the live database file's size.
	SizeBytes() (int64, error)
	// LastBackup is when this process last wrote a backup
	// successfully, or the zero time if it has not.
	LastBackup() time.Time
}

var (
	dbSizeDesc = prometheus.NewDesc("cmdctrl_db_size_bytes",
		"Size of the live SQLite database file.", nil, nil)
	dbBackupDesc = prometheus.NewDesc("cmdctrl_db_backup_last_success_timestamp_seconds",
		"Unix time of the in-process VACUUM INTO backup's last success. 0: none since this process started (the sweep writes one at boot unless it is disabled).",
		nil, nil)
)

type dbCollector struct {
	db  DBSource
	log *slog.Logger
}

// NewDBCollector returns the database collector. log may be nil.
func NewDBCollector(db DBSource, log *slog.Logger) prometheus.Collector {
	if log == nil {
		log = slog.Default()
	}
	return &dbCollector{db: db, log: log}
}

func (c *dbCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- dbSizeDesc
	ch <- dbBackupDesc
}

func (c *dbCollector) Collect(ch chan<- prometheus.Metric) {
	if n, err := c.db.SizeBytes(); err != nil {
		c.log.Warn("metrics: reading the database size failed", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(dbSizeDesc, prometheus.GaugeValue, float64(n))
	}
	var ts float64
	if t := c.db.LastBackup(); !t.IsZero() {
		ts = float64(t.UnixMilli()) / 1000
	}
	ch <- prometheus.MustNewConstMetric(dbBackupDesc, prometheus.GaugeValue, ts)
}
