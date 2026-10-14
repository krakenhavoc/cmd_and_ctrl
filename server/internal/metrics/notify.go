package metrics

import (
	"slices"

	"github.com/prometheus/client_golang/prometheus"
)

// notify.go is ADR 0162 §5's one counter: every Discord DM notification
// the server tried to send, by what it was about and how it went.
//
//	cmdctrl_dm_notifications_total{kind, result}

// Kinds of cmdctrl_dm_notifications_total: the three switches.
const (
	DMKindTurn        = "turn"
	DMKindTableReady  = "table_ready"
	DMKindTableFilled = "table_filled"
)

// Results of cmdctrl_dm_notifications_total.
const (
	// DMSent: Discord accepted the message.
	DMSent = "sent"
	// DMRefused: Discord answered 50007, and the person gets no more
	// until they turn a switch on again.
	DMRefused = "refused"
	// DMError: anything else, a second 429 included. Dropped, never
	// retried late.
	DMError = "error"
)

var (
	dmKindLabels   = []string{DMKindTurn, DMKindTableReady, DMKindTableFilled}
	dmResultLabels = []string{DMSent, DMRefused, DMError}

	dmNotifications = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_dm_notifications_total",
		Help: "Discord DM notifications the server tried to send (ADR 0162), by kind and result.",
	}, []string{"kind", "result"})
)

func init() {
	for _, k := range dmKindLabels {
		for _, r := range dmResultLabels {
			dmNotifications.WithLabelValues(k, r)
		}
	}
}

// DMNotification counts one attempted notification. A kind or result
// outside the closed sets is not counted.
func DMNotification(kind, result string) {
	if !slices.Contains(dmKindLabels, kind) || !slices.Contains(dmResultLabels, result) {
		return
	}
	dmNotifications.WithLabelValues(kind, result).Inc()
}
