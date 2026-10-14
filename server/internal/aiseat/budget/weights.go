package budget

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// EnvPrices overrides or extends the weight table (#3476). The value is
// a comma-separated list of family=in/out/cache_read/cache_write, each
// number a price in US dollars per million tokens, for example
//
//	CMDCTRL_BOT_MODEL_PRICES="haiku=1/5/0.1/1.25,gpt-5=1.25/10/0.125/1.25"
//
// The key is a case-insensitive substring of the model id; the first
// listed key that matches wins, then the built-in families. Unset keeps
// the defaults. A malformed value fails the boot.
const EnvPrices = "CMDCTRL_BOT_MODEL_PRICES"

// Price is what one million tokens of each class weigh, in the budget's
// unit. The unit is the US dollar of the provider's public list price
// per million tokens, so one weighted token is one micro-dollar and a
// cap of 5,000,000 is five dollars.
//
// It is a weight, not a bill: list prices change, batch and priority
// tiers differ, and the code cannot see an invoice. What it buys is
// that a million output tokens of the strongest model no longer count
// the same as a million cache reads of the cheapest.
type Price struct {
	Input, Output, CacheRead, CacheWrite float64
}

// Weigh is the weighted token count of one response's usage.
func (p Price) Weigh(u model.Usage) int64 {
	w := float64(u.InputTokens)*p.Input +
		float64(u.OutputTokens)*p.Output +
		float64(u.CacheReadTokens)*p.CacheRead +
		float64(u.CacheWriteTokens)*p.CacheWrite
	return int64(math.Ceil(w))
}

// family is one row of a Weights table.
type family struct {
	match string
	price Price
}

// Weights maps a model id to its Price.
type Weights struct {
	rows []family
	// fallback prices a model no row matches.
	fallback Price
}

// Anthropic list prices, $ per million tokens: a cache read is a tenth
// of the input price and a five-minute cache write is 1.25 times it.
// Checked against the public pricing page when written (2026-10); the
// default is therefore a weight to refresh, not a promise.
var (
	priceHaiku  = Price{Input: 1, Output: 5, CacheRead: 0.10, CacheWrite: 1.25}
	priceSonnet = Price{Input: 3, Output: 15, CacheRead: 0.30, CacheWrite: 3.75}
	priceOpus   = Price{Input: 5, Output: 25, CacheRead: 0.50, CacheWrite: 6.25}
)

// DefaultWeights prices the three Anthropic families by substring of
// the model id. A model it does not recognise is priced as the most
// expensive family: a cap that under-counts an unknown model fails
// silently in the worse direction.
func DefaultWeights() Weights {
	return Weights{
		rows: []family{
			{"haiku", priceHaiku},
			{"sonnet", priceSonnet},
			{"opus", priceOpus},
		},
		fallback: priceOpus,
	}
}

// For returns the price of a model id.
func (w Weights) For(modelID string) Price {
	id := strings.ToLower(modelID)
	for _, r := range w.rows {
		if strings.Contains(id, r.match) {
			return r.price
		}
	}
	return w.fallback
}

// WeightsFromEnv is DefaultWeights with CMDCTRL_BOT_MODEL_PRICES
// applied in front.
func WeightsFromEnv() (Weights, error) {
	return weightsFrom(os.Getenv)
}

func weightsFrom(getenv func(string) string) (Weights, error) {
	w := DefaultWeights()
	raw := strings.TrimSpace(getenv(EnvPrices))
	if raw == "" {
		return w, nil
	}
	var rows []family
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key, nums, ok := strings.Cut(item, "=")
		key = strings.ToLower(strings.TrimSpace(key))
		parts := strings.Split(nums, "/")
		if !ok || key == "" || len(parts) != 4 {
			return Weights{}, fmt.Errorf("%s: %q: want family=in/out/cache_read/cache_write", EnvPrices, item)
		}
		var v [4]float64
		for i, s := range parts {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil || f < 0 || math.IsNaN(f) || math.IsInf(f, 0) {
				return Weights{}, fmt.Errorf("%s: %q: %q is not a non-negative price", EnvPrices, item, s)
			}
			v[i] = f
		}
		rows = append(rows, family{key, Price{v[0], v[1], v[2], v[3]}})
	}
	w.rows = append(rows, w.rows...)
	return w, nil
}

// FlatWeights counts every token class of every model at 1, which is
// the arithmetic before #3476. For tests.
func FlatWeights() Weights {
	flat := Price{1, 1, 1, 1}
	return Weights{fallback: flat}
}
