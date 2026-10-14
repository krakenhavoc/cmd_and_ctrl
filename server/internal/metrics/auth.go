package metrics

import "github.com/prometheus/client_golang/prometheus"

// Reasons of cmdctrl_auth_failures_total: the sign-in failures a player
// meets that no HTTP status shows. The role gate fails closed with a
// redirect (a 302), so cmdctrl_http_requests_total never sees it.
const (
	// AuthExchange: Discord refused or failed the OAuth code exchange.
	AuthExchange = "oauth_exchange"
	// AuthProfile: Discord would not return the signed-in user.
	AuthProfile = "oauth_profile"
	// AuthGateCheck: Discord did not answer the sign-up gate's role
	// check, so the sign-in was refused (fail closed).
	AuthGateCheck = "gate_check"
	// AuthRecord: the sign-in could not be recorded (a database error).
	AuthRecord = "signin_record"
)

var authFailureReasons = []string{AuthExchange, AuthProfile, AuthGateCheck, AuthRecord}

var authFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "cmdctrl_auth_failures_total",
	Help: "Discord sign-in attempts that failed on the server's side or Discord's, by reason. A person who simply lacks the role is not a failure and is not counted.",
}, []string{"reason"})

func init() {
	for _, r := range authFailureReasons {
		authFailures.WithLabelValues(r)
	}
}

// AuthFailed counts one failed sign-in, by reason (an Auth* value).
func AuthFailed(reason string) { authFailures.WithLabelValues(reason).Inc() }
