package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// bots.go is ADR 0123 §3's "Bot and model seats" table: totals across
// every bot runner, recorded where the runner and the model funnel
// already count (aiseat.Runner's decision counters, aiseat/model's
// recorder), never by polling a runner's Stats.

// Values of the tier label: aiseat's four tiers, and "other" for a
// policy that is none of them (a test or harness policy).
const (
	BotTierRandom    = "random"
	BotTierHeuristic = "heuristic"
	BotTierAssisted  = "assisted"
	BotTierStrong    = "strong"
	BotTierOther     = "other"
)

// Values of the layer label: which layer answered a decision (ADR
// 0033 §5's A, B and C), and "random" for the random tier, which has
// no layers.
const (
	BotLayerA      = "A"
	BotLayerB      = "B"
	BotLayerC      = "C"
	BotLayerRandom = "random"
)

// Values of the result label on cmdctrl_bot_model_calls_total. The
// action metrics' applied and rejected share the label name.
const (
	ModelResultOK         = "ok"
	ModelResultTimeout    = "timeout"
	ModelResultError      = "error"
	ModelResultMalformed  = "malformed"
	ModelResultOutOfRange = "out_of_range"
)

// Values of the direction label on cmdctrl_bot_model_tokens_total.
const (
	TokensPrompt     = "prompt"
	TokensCompletion = "completion"
)

// BotFallbackOther is the cause of a fallback outside the closed set
// below: a cause added to aiseat without its row here.
const BotFallbackOther = "other"

var (
	botTierLabels     = []string{BotTierRandom, BotTierHeuristic, BotTierAssisted, BotTierStrong, BotTierOther}
	botLayerLabels    = []string{BotLayerA, BotLayerB, BotLayerC, BotLayerRandom}
	modelResultLabels = []string{ModelResultOK, ModelResultTimeout, ModelResultError, ModelResultMalformed, ModelResultOutOfRange}
	directionLabels   = []string{TokensPrompt, TokensCompletion}

	// runnerFallbackCauses are aiseat's runner fallback causes
	// (aiseat.Fallback*): why the answer the runner took was not the
	// one the policy returned. The decision log writes them as
	// runner_fallback.
	runnerFallbackCauses = []string{"timeout", "policy-error", "out-of-range", "decline-pass", "decline-always-legal"}

	// modelFallbackCauses maps the model funnel's causes
	// (aiseat/model.Fallback*, the decision log's trace.fallback) to
	// their label: the same word with "model-" in front, because two
	// of them (out-of-range, policy-error) are spelled like a runner
	// cause and mean something else.
	modelFallbackCauses = map[string]string{
		"no-client":    "model-no-client",
		"no-budget":    "model-no-budget",
		"error":        "model-error",
		"malformed":    "model-malformed",
		"out-of-range": "model-out-of-range",
		"policy-error": "model-policy-error",
	}
)

// botFallbackCauseLabels is the closed set of the cause label.
func botFallbackCauseLabels() []string {
	out := append([]string{BotFallbackOther}, runnerFallbackCauses...)
	for _, v := range modelFallbackCauses {
		out = append(out, v)
	}
	return out
}

// decisionBuckets reach past the slowest tier's deadline: 2 s by
// default, 5 s for strong, 20 s with a self-hosted model, 120 s with
// thinking on.
var decisionBuckets = []float64{.001, .01, .05, .1, .25, .5, 1, 2, 5, 10, 20, 30, 60, 120}

var (
	botDecisions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_bot_decisions_total",
		Help: "Bot decisions that produced a move, by tier and the layer that answered (A rules, B heuristic, C model, or random).",
	}, []string{"tier", "layer"})

	botDecisionSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cmdctrl_bot_decision_seconds",
		Help:    "Time a bot policy took to decide one window, deadline misses included.",
		Buckets: decisionBuckets,
	}, []string{"tier"})

	botFallbacks = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_bot_fallbacks_total",
		Help: "Bot windows whose answer fell back: the runner's causes as the decision log names them, and the model funnel's prefixed model-.",
	}, []string{"tier", "cause"})

	botModelCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_bot_model_calls_total",
		Help: "Decision calls a model-backed bot made, by how they ended.",
	}, []string{"tier", "result"})

	botModelCallSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cmdctrl_bot_model_call_seconds",
		Help:    "Time inside one bot decision call to a model.",
		Buckets: decisionBuckets,
	}, []string{"tier"})

	botModelTokens = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_bot_model_tokens_total",
		Help: "Tokens the bot model calls reported, decisions and improvisations together: prompt (input plus cache reads and writes) and completion.",
	}, []string{"tier", "direction"})
)

// BotTierLabel is the tier label for a policy name: one of the four
// tiers, or "other".
func BotTierLabel(policy string) string {
	switch policy {
	case BotTierRandom, BotTierHeuristic, BotTierAssisted, BotTierStrong:
		return policy
	}
	return BotTierOther
}

// IsBotLayer reports whether layer is a value of the layer label.
func IsBotLayer(layer string) bool {
	switch layer {
	case BotLayerA, BotLayerB, BotLayerC, BotLayerRandom:
		return true
	}
	return false
}

// RecordBotDecision counts one decision. tier is a BotTierLabel; a
// layer outside the set counts as B, the heuristic, which is the one
// layer that does not name itself.
func RecordBotDecision(tier, layer string) {
	if !IsBotLayer(layer) {
		layer = BotLayerB
	}
	botDecisions.WithLabelValues(BotTierLabel(tier), layer).Inc()
}

// ObserveBotDecision observes one window's decision time.
func ObserveBotDecision(tier string, d time.Duration) {
	botDecisionSeconds.WithLabelValues(BotTierLabel(tier)).Observe(d.Seconds())
}

// RecordRunnerFallback counts one of the runner's fallbacks (an
// aiseat.Fallback* cause).
func RecordRunnerFallback(tier, cause string) {
	label := BotFallbackOther
	for _, c := range runnerFallbackCauses {
		if c == cause {
			label = c
			break
		}
	}
	botFallbacks.WithLabelValues(BotTierLabel(tier), label).Inc()
}

// RecordModelFallback counts one of the model funnel's fallbacks (an
// aiseat/model.Fallback* cause), labelled with "model-" in front.
func RecordModelFallback(tier, cause string) {
	label, ok := modelFallbackCauses[cause]
	if !ok {
		label = BotFallbackOther
	}
	botFallbacks.WithLabelValues(BotTierLabel(tier), label).Inc()
}

// RecordBotModelCall counts one decision call to a model and observes
// its latency. result is a ModelResult value.
func RecordBotModelCall(tier, result string, d time.Duration) {
	switch result {
	case ModelResultOK, ModelResultTimeout, ModelResultError, ModelResultMalformed, ModelResultOutOfRange:
	default:
		result = ModelResultError
	}
	tier = BotTierLabel(tier)
	botModelCalls.WithLabelValues(tier, result).Inc()
	botModelCallSeconds.WithLabelValues(tier).Observe(d.Seconds())
}

// AddBotModelTokens adds what one model call reported.
func AddBotModelTokens(tier string, prompt, completion int) {
	tier = BotTierLabel(tier)
	if prompt > 0 {
		botModelTokens.WithLabelValues(tier, TokensPrompt).Add(float64(prompt))
	}
	if completion > 0 {
		botModelTokens.WithLabelValues(tier, TokensCompletion).Add(float64(completion))
	}
}
