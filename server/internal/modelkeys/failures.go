package modelkeys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/netguard"
)

// The fixed sentences. Every failure a person, a log or a table sees is
// one of these: a provider's own text can echo part of the key, and a
// transport error can name an address the guard refused.
const (
	ReasonRefused     = "The provider refused the key. Check it, or paste a new one."
	ReasonQuota       = "The provider says the key is out of credit or over its quota."
	ReasonModel       = "The provider does not offer that model to this key."
	ReasonUnreadable  = "The saved key could not be read. Paste it again."
	ReasonPrivate     = "The provider's address is on a private or local network, which this server will not contact."
	ReasonUnreachable = "Could not reach the provider."
	ReasonTimeout     = "The provider did not answer in time."
	ReasonRateLimited = "The provider is limiting requests; it should pass in a minute."
	ReasonBadReply    = "The provider's reply was not understood."
	ReasonRemoved     = "The key was removed."
	ReasonDailyLimit  = "The key's games-a-day limit is used up for today."
	ReasonCallLimit   = "The key's calls-per-game limit is used up for this game."
)

// classify sorts a provider call's failure. Fatal means the key will
// not work for the rest of this game (refused, out of quota, no such
// model, an address the guard refuses), so the seat stops calling and
// the owner is told; anything else is one window's fallback.
func classify(err error) (fatal bool, reason string) {
	var api *model.APIError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return false, ReasonTimeout
	case errors.As(err, &api):
		text := strings.ToLower(api.Kind + " " + api.Message)
		switch {
		case api.Status == http.StatusUnauthorized, api.Status == http.StatusForbidden:
			return true, ReasonRefused
		case api.Status == http.StatusPaymentRequired,
			strings.Contains(text, "insufficient_quota"), strings.Contains(text, "credit balance"),
			api.Status == http.StatusTooManyRequests && strings.Contains(text, "quota"):
			return true, ReasonQuota
		case api.Status == http.StatusNotFound, strings.Contains(text, "model_not_found"):
			return true, ReasonModel
		case api.Status == http.StatusTooManyRequests:
			return false, ReasonRateLimited
		case api.Status >= 300 && api.Status < 400:
			// A redirect is never followed (netguard); an endpoint that
			// answers with one will answer with one every time.
			return true, ReasonUnreachable
		}
		return false, fmt.Sprintf("The provider answered with an error (status %d).", api.Status)
	case strings.Contains(err.Error(), netguard.ErrAddressRefused.Error()):
		// The model clients flatten the transport error into ErrOutage's
		// text, so the guard's refusal is recognised by its sentence.
		return true, ReasonPrivate
	case errors.Is(err, model.ErrOutage):
		return false, ReasonUnreachable
	}
	return false, ReasonBadReply
}

// testFailure is the sentence the Settings test shows.
func testFailure(err error) string {
	_, reason := classify(err)
	return reason
}

// sanitize turns a provider call's error into one the policy can log:
// the same kind (a deadline is still a deadline, an HTTP status is
// still a status) with none of the provider's or the transport's text.
func sanitize(err error, reason string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	case errors.Is(err, context.Canceled):
		return context.Canceled
	}
	var api *model.APIError
	if errors.As(err, &api) {
		return &model.APIError{Provider: "own-key", Status: api.Status, Message: reason}
	}
	if errors.Is(err, model.ErrOutage) {
		return fmt.Errorf("%w: own key: %s", model.ErrOutage, reason)
	}
	return errors.New("own key: " + reason)
}

// reasonCode is the short metrics code (metrics.OwnKey*) for one of the
// fixed sentences above; anything else is "other".
func reasonCode(reason string) string {
	switch reason {
	case ReasonRefused:
		return metrics.OwnKeyRefused
	case ReasonQuota:
		return metrics.OwnKeyQuota
	case ReasonModel:
		return metrics.OwnKeyModel
	case ReasonUnreadable:
		return metrics.OwnKeyUnreadable
	case ReasonPrivate:
		return metrics.OwnKeyPrivate
	case ReasonUnreachable:
		return metrics.OwnKeyUnreachable
	case ReasonTimeout:
		return metrics.OwnKeyTimeout
	case ReasonRateLimited:
		return metrics.OwnKeyRateLimited
	case ReasonBadReply:
		return metrics.OwnKeyBadReply
	case ReasonRemoved:
		return metrics.OwnKeyRemoved
	case ReasonDailyLimit:
		return metrics.OwnKeyDailyLimit
	case ReasonCallLimit:
		return metrics.OwnKeyCallLimit
	}
	return metrics.OwnKeyOther
}
