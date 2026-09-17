package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/deckprofile"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // catalog hooks
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// probe.go tests two transport hypotheses that cannot be tested
// offline, and that between them would make a `qwen3:14b` seat play
// the heuristic on every single window while reporting itself as
// `assisted`:
//
//  1. **Silent truncation.** Ollama defaults to a small context on a
//     modest card and drops the FRONT of an over-long prompt with no
//     error. The front is the rules primer and the instructions. The
//     OpenAI-compatible endpoint has no field to set num_ctx with, so
//     the only evidence available to a client is usage.prompt_tokens
//     coming back far below what was sent.
//  2. **Thinking not suppressed.** A thinking model inside a 128-token
//     cap produces an empty `content`, `finish_reason: "length"`, and
//     the whole reply in `message.reasoning` — which parses as
//     nothing, scores as FallbackMalformed, and plays Layer B. #839
//     measured that Ollama's `/v1` endpoint ignores `think:false` and
//     honours `reasoning_effort:"none"`, and the client now sends
//     both, so this verdict should come back SUPPRESSED. It is kept
//     because a different server may honour neither, and because a
//     regression here is invisible from the table: the seat still
//     plays, it just plays Layer B under a model tier's name.
//
// The probe sends ONE request through the ordinary OpenAIClient, in
// the exact shape the funnel sends, and prints the evidence. It always
// exits 0: it is a diagnostic, and "the endpoint is down" is a
// finding, not a failure.

// estimateDivisor turns prompt bytes into a rough token count. Three
// bytes per token is conservative for English prose with a lot of
// punctuation and card names, which is what this prompt is; it is a
// yardstick for "did roughly what I sent arrive", not an accounting.
const estimateDivisor = 3

// truncationRatio is how far below the estimate prompt_tokens has to
// land before truncation is the likely explanation. Tokenisers vary
// by a lot less than 40%.
const truncationRatio = 0.6

// probeResult is everything one call produced, and the only input to
// the verdicts. It is a plain struct so the verdict logic can be
// tested without an endpoint.
type probeResult struct {
	Endpoint    string
	Model       string
	SystemBytes int
	UserBytes   int

	PromptTokens       int
	CompletionTokens   int
	CachedPromptTokens int
	FinishReason       string
	Reply              string
	Reasoning          string
	ParsedIndex        int
	ParseErr           error
	Moves              int
	Elapsed            time.Duration
	CallErr            error
}

// Estimate is the client-side guess at how many tokens were sent.
func (r probeResult) Estimate() int {
	return (r.SystemBytes + r.UserBytes) / estimateDivisor
}

// ReplyParsed reports whether the model answered in the shape it was
// asked for.
func (r probeResult) ReplyParsed() bool {
	if strings.TrimSpace(r.Reply) == "" || r.ParseErr != nil {
		return false
	}
	return r.ParsedIndex >= 0 && r.ParsedIndex < r.Moves
}

// thinkingEvidence reports whether the reply's shape is already
// explained by a model that thought instead of answering.
//
// It exists to keep the two verdicts from both firing on one piece of
// evidence. An unparseable reply is weak evidence of truncation and
// STRONG evidence of unsuppressed thinking, and on the measured
// qwen3 shape — empty content, finish_reason "length", the whole
// answer in message.reasoning — the prompt arrived intact. Reporting
// that as "TRUNCATION: likely" would send an operator to raise a
// context length that was never the problem.
func thinkingEvidence(r probeResult) bool {
	return strings.TrimSpace(r.Reasoning) != "" ||
		(r.FinishReason == "length" && strings.TrimSpace(r.Reply) == "")
}

// truncationVerdict answers hypothesis 1.
//
// Two independent signatures: the server counted far fewer prompt
// tokens than were sent (the front of the prompt was dropped), or the
// model answered in a shape the instructions forbid — which is what a
// model that never saw the instructions does. The second is only
// reported when thinking does not already explain it; see
// thinkingEvidence.
func truncationVerdict(r probeResult) string {
	if r.CallErr != nil {
		return "TRUNCATION: unknown — the call did not return"
	}
	est := r.Estimate()
	switch {
	case r.PromptTokens == 0:
		return "TRUNCATION: unknown — the endpoint reported no prompt_tokens"
	case float64(r.PromptTokens) < float64(est)*truncationRatio:
		return fmt.Sprintf("TRUNCATION: likely — the endpoint counted %d prompt tokens for ~%d sent (<%.0f%%); raise the server's context (OLLAMA_CONTEXT_LENGTH) or shrink the prompt",
			r.PromptTokens, est, truncationRatio*100)
	case !r.ReplyParsed() && thinkingEvidence(r):
		return fmt.Sprintf("TRUNCATION: not detected — %d prompt tokens for ~%d sent. The reply is unusable, but thinking explains that (see below), not a truncated prompt.",
			r.PromptTokens, est)
	case !r.ReplyParsed():
		return "TRUNCATION: likely — the prompt token count is plausible, but the reply ignores the answer format, which is what a model that never saw the instructions does"
	default:
		return fmt.Sprintf("TRUNCATION: not detected — %d prompt tokens for ~%d sent, and the reply is in the requested shape", r.PromptTokens, est)
	}
}

// thinkingVerdict answers hypothesis 2.
//
// A non-empty reasoning field is direct evidence. So is the
// signature of a thinking model that ran out of budget before it got
// to the answer: finish_reason "length" with nothing usable in
// content — the shape Ollama 0.34 with qwen3:14b returned before #839
// added `reasoning_effort:"none"`, and the shape to watch for on any
// server that honours neither switch.
func thinkingVerdict(r probeResult) string {
	if r.CallErr != nil {
		return "THINKING: unknown — the call did not return"
	}
	switch {
	case strings.TrimSpace(r.Reasoning) != "":
		return fmt.Sprintf("THINKING: not suppressed — the reply carried a reasoning field (%d chars). This server honours neither `think:false` nor `reasoning_effort:\"none\"`; every window on it will score as malformed and play the heuristic.",
			len(r.Reasoning))
	case thinkingEvidence(r):
		return "THINKING: not suppressed — finish_reason is \"length\" and the reply is empty or unparseable: the budget went somewhere that is not the answer"
	default:
		return "THINKING: suppressed (or this model does not think) — no reasoning field and the answer arrived inside the token cap"
	}
}

func runProbe(args []string) int {
	fs := flag.NewFlagSet("probe", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	endpoint := fs.String("endpoint", "", "OpenAI-compatible endpoint (default: $CMDCTRL_OPENAI_ENDPOINT)")
	modelID := fs.String("model", "", "model id (default: $CMDCTRL_BOT_MODEL)")
	maxTokens := fs.Int("max-tokens", 0, "max_tokens for the call (default: the assisted tier's routine profile)")
	deckID := fs.String("deck", "izzet-aggro", "curated deck whose list becomes the static prompt block")
	dump := fs.String("dump", "", "Scryfall bulk dump for oracle text (default: $CMDCTRL_SCRYFALL_DUMP)")
	timeout := fs.Duration("timeout", 5*time.Minute, "wall clock for the one call")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	out := os.Stdout

	client, url := buildClient(*endpoint)
	if client == nil {
		say(out, "%s\n", "no endpoint configured: pass --endpoint or set CMDCTRL_OPENAI_ENDPOINT")
		return 0
	}
	id := strings.TrimSpace(*modelID)
	if id == "" {
		id = strings.TrimSpace(os.Getenv("CMDCTRL_BOT_MODEL"))
	}
	if id == "" {
		say(out, "%s\n", "no model configured: pass --model or set CMDCTRL_BOT_MODEL")
		return 0
	}

	idx := loadIndex(out, *dump)
	profile, profileSource := buildProfile(idx, *deckID)

	cfg := model.DefaultConfig()
	cfg.Deck = profile
	cfg.Routine.ID, cfg.Frontier.ID = id, id
	if *maxTokens > 0 {
		cfg.Routine.MaxTokens, cfg.Frontier.MaxTokens = *maxTokens, *maxTokens
	}
	pol := model.New(cfg)

	in, err := representativeInput(idx, *deckID)
	if err != nil {
		say(out, "could not build a representative position: %v\n", err)
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	req, _, _ := pol.BuildRequest(ctx, in)

	res := probeResult{
		Endpoint:    url,
		Model:       id,
		SystemBytes: systemBytes(req),
		UserBytes:   len(req.User),
		Moves:       len(in.Moves),
	}
	started := time.Now()
	resp, cerr := client.Complete(ctx, req)
	res.Elapsed = time.Since(started)
	res.CallErr = cerr
	if cerr == nil {
		res.PromptTokens = resp.Usage.InputTokens
		res.CompletionTokens = resp.Usage.OutputTokens
		res.CachedPromptTokens = resp.Usage.CachedPromptTokens
		res.FinishReason = resp.StopReason
		res.Reply = resp.Text
		res.Reasoning = resp.Reasoning
		res.ParsedIndex, res.ParseErr = model.ParseAnswerIndex(resp.Text)
	}

	printProbe(out, res, profileSource, in)
	return 0
}

// buildClient applies model.NewOpenAIClient's env semantics with the
// flag taking precedence.
func buildClient(endpoint string) (*model.OpenAIClient, string) {
	if e := strings.TrimSpace(endpoint); e != "" {
		// Set the variable rather than constructing the client by
		// hand, so the probe inherits every rule NewOpenAIClient has
		// about completing a bare host and reading the key and the
		// think switch. The probe is a short-lived process.
		_ = os.Setenv(model.EnvOpenAIEndpoint, e)
	}
	c := model.NewOpenAIClient()
	if c == nil {
		return nil, ""
	}
	return c, c.URL()
}

func loadIndex(out io.Writer, dump string) *cards.Index {
	path := strings.TrimSpace(dump)
	if path == "" {
		path = strings.TrimSpace(os.Getenv("CMDCTRL_SCRYFALL_DUMP"))
	}
	if path == "" {
		return nil
	}
	idx := cards.NewIndex()
	n, err := idx.Load(path)
	if err != nil {
		say(out, "scryfall dump %s did not load (%v); the prompt will carry card NAMES only\n", path, err)
		return nil
	}
	say(out, "scryfall dump: %s (%d cards)\n", path, n)
	return idx
}

// buildProfile is the static half of the prompt: the real deck
// profile when a dump is loaded, a name-only one otherwise. The
// difference is thousands of tokens, so the probe says which it used.
func buildProfile(idx *cards.Index, deckID string) (model.DeckProfile, string) {
	if p, ok := deckprofile.Build(idx, deckID); ok {
		if idx == nil {
			return p, "names only (no Scryfall dump)"
		}
		return p, "full decklist with oracle text"
	}
	d, ok := decks.Lookup(deckID)
	if !ok {
		return model.DeckProfile{}, "none (unknown deck)"
	}
	return model.DeckProfile{Name: d.Name}, "deck name only"
}

func systemBytes(req model.Request) int {
	n := 0
	for _, b := range req.System {
		n += len(b.Text)
	}
	return n
}

// representativeInput plays a real four-seat game forward with the
// heuristic in every chair and stops at the first window that Layer A
// did NOT absorb and that has a few moves in it — i.e. the kind of
// window a model tier is actually asked about.
//
// A synthetic Input would be quicker and would measure the wrong
// thing: the whole question is how big a real prompt is, and a real
// prompt has a real board in it.
func representativeInput(idx *cards.Index, deckID string) (aiseat.Input, error) {
	const seats = 4
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), seatDeck(idx, deckID)); err != nil {
			return aiseat.Input{}, fmt.Errorf("add player: %w", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(31, 32))); err != nil {
		return aiseat.Input{}, fmt.Errorf("start: %w", err)
	}

	pol := heuristic.New()
	ctx := context.Background()
	ids := make([]uuid.UUID, 0, seats)
	for _, p := range g.Seats {
		ids = append(ids, p.ID)
	}
	for step := 0; step < 20000; step++ {
		if g.CurrentState() != game.StateActive {
			break
		}
		acted := false
		for _, id := range ids {
			moves := legal.EnumerateFor(g, id)
			if len(moves) == 0 {
				continue
			}
			in := aiseat.Input{
				View:  protocol.ViewOfGameFor(g, id.String()),
				Seat:  id,
				Moves: moves,
			}
			// The window worth probing with: past the mulligans, not
			// settled by Layer A, and with enough on offer that the
			// move list is worth rendering.
			if in.View.Turn.Number >= 4 && len(moves) >= 5 && !rules.Resolve(in).Absorbed() {
				return in, nil
			}
			d, err := pol.Decide(ctx, in)
			if err != nil || d.Index < 0 || d.Index >= len(moves) {
				if pi := aiseat.PassIndex(moves); pi >= 0 {
					d.Index = pi
				} else {
					d.Index = 0
				}
			}
			mv := moves[d.Index]
			if derr := actions.Dispatch(g, actions.Action{
				Type:   actions.Type(mv.Type),
				Player: mv.Player,
				Caller: id,
				Params: mv.Params,
			}); derr == nil {
				acted = true
			}
		}
		if !acted {
			break
		}
	}
	return aiseat.Input{}, fmt.Errorf("no escalated window turned up in 20000 steps")
}

// seatDeck is the curated deck when a dump resolved it, and a vanilla
// battle deck otherwise — the probe must work on a laptop with no
// 600 MiB dump on it. A deck that will not resolve is not an error
// here: a thinner board still measures the transport.
func seatDeck(idx *cards.Index, deckID string) []game.Card {
	if idx != nil {
		if list, err := decks.Load(idx, deckID); err == nil {
			return list.ToGameCards()
		}
	}
	return vanillaDeck()
}

func vanillaDeck() []game.Card {
	var deck []game.Card
	cmdr := game.NewCommander("Commander Bear", uuid.Nil)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	add := func(n int, build func() game.Card) {
		for i := 0; i < n; i++ {
			deck = append(deck, build())
		}
	}
	add(26, func() game.Card {
		c := game.NewCard("Mountain", uuid.Nil)
		c.TypeLine = "Basic Land — Mountain"
		return c
	})
	add(16, func() game.Card {
		c := game.NewCard("Bear", uuid.Nil)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		return c
	})
	add(12, func() game.Card {
		c := game.NewCard("Drake", uuid.Nil)
		c.TypeLine = "Creature — Drake"
		c.ManaCost = "{2}{R}"
		c.Power, c.Toughness = 3, 3
		c.Keywords = []string{"flying"}
		return c
	})
	add(10, func() game.Card {
		c := game.NewCard("Ogre", uuid.Nil)
		c.TypeLine = "Creature — Ogre"
		c.ManaCost = "{3}{R}"
		c.Power, c.Toughness = 4, 4
		return c
	})
	return deck
}

func printProbe(out io.Writer, r probeResult, profileSource string, in aiseat.Input) {
	say(out, "%s\n", "--- boteval probe ------------------------------------------")
	say(out, "endpoint           %s\n", r.Endpoint)
	say(out, "model              %s\n", r.Model)
	say(out, "deck profile       %s\n", profileSource)
	say(out, "window             turn %d, %s, %d legal moves\n",
		in.View.Turn.Number, in.View.Turn.Step, len(in.Moves))
	say(out, "system bytes       %d\n", r.SystemBytes)
	say(out, "user bytes         %d\n", r.UserBytes)
	say(out, "estimate (b/%d)     ~%d tokens\n", estimateDivisor, r.Estimate())
	if r.CallErr != nil {
		say(out, "call               FAILED after %v: %v\n", r.Elapsed.Round(time.Millisecond), r.CallErr)
		say(out, "%s\n", "------------------------------------------------------------")
		say(out, "%s\n", truncationVerdict(r))
		say(out, "%s\n", thinkingVerdict(r))
		return
	}
	say(out, "prompt_tokens      %d\n", r.PromptTokens)
	say(out, "completion_tokens  %d\n", r.CompletionTokens)
	say(out, "cached_tokens      %d (server-side prefix cache)\n", r.CachedPromptTokens)
	say(out, "finish_reason      %q\n", r.FinishReason)
	say(out, "reply empty        %v\n", strings.TrimSpace(r.Reply) == "")
	say(out, "reasoning field    %v (%d chars)\n", strings.TrimSpace(r.Reasoning) != "", len(r.Reasoning))
	say(out, "wall time          %v\n", r.Elapsed.Round(time.Millisecond))
	say(out, "reply (first 300)  %s\n", model.OneLine(model.Truncate(r.Reply, 300)))
	if strings.TrimSpace(r.Reasoning) != "" {
		say(out, "reasoning (first 300) %s\n", model.OneLine(model.Truncate(r.Reasoning, 300)))
	}
	if r.ParseErr != nil {
		say(out, "parsed index       PARSE FAILED: %v\n", r.ParseErr)
	} else {
		say(out, "parsed index       %d (of %d moves; in range: %v)\n", r.ParsedIndex, r.Moves, r.ReplyParsed())
	}
	say(out, "%s\n", "------------------------------------------------------------")
	say(out, "%s\n", truncationVerdict(r))
	say(out, "%s\n", thinkingVerdict(r))
}

// say writes one line to the probe's output. The error is
// deliberately dropped: this is a diagnostic printing to a terminal,
// and a write that fails has nowhere left to report it.
func say(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
