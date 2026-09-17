package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/suite"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

// suite.go is `boteval suite`: run a labelled position suite against
// a policy, harvest candidate positions out of decision logs, and
// render one position as the prompt a model would see.
//
// The heuristic half of this runs in CI as an ordinary Go test
// (aiseat/suite's own suite_test.go) because it is pure computation
// over frozen inputs. This binary is for the half CI cannot do: a
// model tier needs an endpoint, a Scryfall dump and minutes of wall
// clock, and its output is a report somebody reads.

// defaultPositionsDir is where the committed suite lives, relative to
// `server/`. Every boteval invocation in the docs is run from there.
const defaultPositionsDir = "internal/aiseat/suite/testdata/positions"

func runSuite(args []string) int {
	if len(args) == 0 {
		suiteUsage()
		return 2
	}
	switch args[0] {
	case "run":
		return runSuiteRun(args[1:])
	case "harvest":
		return runSuiteHarvest(args[1:])
	case "render":
		return runSuiteRender(args[1:])
	case "-h", "--help", "help":
		suiteUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "boteval suite: unknown verb %q\n\n", args[0])
		suiteUsage()
		return 2
	}
}

func suiteUsage() {
	fmt.Fprint(os.Stderr, `boteval suite — the labelled position suite

  suite run      ask a policy every labelled position and report agreement.
                 boteval suite run [--dir DIR] [--policy heuristic|assisted|strong]
                                   [--deck ID] [--max-think 20s] [--parallel 1]
                                   [--out report.json] [--md]

  suite harvest  pull candidate windows out of decision logs into an inbox of
                 UNLABELLED positions, for a human to answer.
                 boteval suite harvest --from 'dir/*.decisions.jsonl' --to inbox/
                                   [--escalated] [--disagree] [--fallback a,b]
                                   [--layer A,B,C] [--seat 0,1] [--tag block,attack]
                                   [--limit N] [--seed N]

  suite render   print the exact prompt a model would see for one position, with
                 the move list annotated. This is the labelling screen.
                 boteval suite render --pos path/to/position.json [--deck ID]

Env fallbacks: CMDCTRL_OPENAI_ENDPOINT, CMDCTRL_OPENAI_API_KEY, CMDCTRL_BOT_MODEL,
CMDCTRL_BOT_MAX_THINK, CMDCTRL_SCRYFALL_DUMP.
`)
}

// --- suite run -------------------------------------------------------

type suiteRunOpts struct {
	Dir      string
	Policy   string
	Deck     string
	Dump     string
	MaxThink time.Duration
	Parallel int
	Out      string
	MD       bool
}

func parseSuiteRun(args []string) (suiteRunOpts, error) {
	var o suiteRunOpts
	fs := flag.NewFlagSet("suite run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&o.Dir, "dir", defaultPositionsDir, "directory of position files")
	fs.StringVar(&o.Policy, "policy", "heuristic", "policy to ask: heuristic, assisted or strong")
	fs.StringVar(&o.Deck, "deck", "izzet-aggro", "curated deck whose list becomes the static prompt block (model tiers only)")
	fs.StringVar(&o.Dump, "dump", "", "Scryfall bulk dump for oracle text (default: $CMDCTRL_SCRYFALL_DUMP)")
	fs.DurationVar(&o.MaxThink, "max-think", 0, "hard deadline per position (default: $CMDCTRL_BOT_MAX_THINK, else the suite's own)")
	fs.IntVar(&o.Parallel, "parallel", 1, "positions in flight at once")
	fs.StringVar(&o.Out, "out", "", "write the report as JSON to this path")
	fs.BoolVar(&o.MD, "md", false, "print the markdown report block on stdout")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if o.Parallel < 1 {
		return o, fmt.Errorf("--parallel must be at least 1, got %d", o.Parallel)
	}
	if _, err := tiers.Parse(o.Policy); err != nil {
		return o, err
	}
	if o.MaxThink <= 0 {
		o.MaxThink = envDuration("CMDCTRL_BOT_MAX_THINK")
	}
	if o.MaxThink <= 0 {
		o.MaxThink = suite.DefaultMaxThink
	}
	return o, nil
}

func runSuiteRun(args []string) int {
	o, err := parseSuiteRun(args)
	if err != nil {
		return 2
	}
	out := os.Stdout

	positions, err := suite.Load(o.Dir)
	if err != nil {
		say(os.Stderr, "%v\n", err)
		return 1
	}
	if len(positions) == 0 {
		say(os.Stderr, "no positions in %s\n", o.Dir)
		return 1
	}

	pol, err := buildSuitePolicy(out, o)
	if err != nil {
		say(os.Stderr, "%v\n", err)
		return 1
	}

	rep := suite.Run(context.Background(), positions, pol, suite.RunOptions{
		MaxThink: o.MaxThink,
		Parallel: o.Parallel,
	})

	if o.MD {
		say(out, "%s\n", rep.Markdown())
	} else {
		printSuiteReport(out, rep)
	}
	if o.Out != "" {
		blob, merr := json.MarshalIndent(rep, "", "  ")
		if merr != nil {
			say(os.Stderr, "marshal report: %v\n", merr)
			return 1
		}
		if werr := os.WriteFile(o.Out, append(blob, '\n'), 0o600); werr != nil {
			say(os.Stderr, "write %s: %v\n", o.Out, werr)
			return 1
		}
		say(out, "report written to %s\n", o.Out)
	}

	// A gate failure is the one thing this command exits non-zero
	// for: somebody pinned this position for this policy, and it
	// stopped passing.
	if fails := rep.GateFailures(rep.Policy); len(fails) > 0 {
		say(os.Stderr, "\n%d GATED position(s) failed for %q:\n", len(fails), rep.Policy)
		for _, f := range fails {
			say(os.Stderr, "  %s\n", f)
		}
		return 1
	}
	return 0
}

// buildSuitePolicy builds the tier the run asked for. A model tier
// with no transport is REFUSED rather than quietly run as Layer A +
// B, which is exactly tiers.Factory.TierStatus's rule and for the
// same reason: a report labelled `assisted` that measured the
// heuristic is a false measurement.
func buildSuitePolicy(out *os.File, o suiteRunOpts) (aiseat.Policy, error) {
	tier, err := tiers.Parse(o.Policy)
	if err != nil {
		return nil, err
	}
	opt := tiers.Options{MaxThink: o.MaxThink}
	if tier.NeedsModel() {
		client, url := buildClient("")
		if client == nil {
			return nil, fmt.Errorf("tier %q needs a model: set CMDCTRL_OPENAI_ENDPOINT to a local LLM (Ollama, LM Studio, llama.cpp, vLLM) or CMDCTRL_ANTHROPIC_API_KEY to a hosted one, then re-run", tier)
		}
		id := strings.TrimSpace(os.Getenv("CMDCTRL_BOT_MODEL"))
		if id == "" {
			return nil, fmt.Errorf("tier %q needs a model id: set CMDCTRL_BOT_MODEL", tier)
		}
		idx := loadIndex(out, o.Dump)
		profile, source := buildProfile(idx, o.Deck)
		say(out, "endpoint %s · model %s · deck profile %s\n", url, id, source)
		opt.Client = client
		opt.Deck = profile
		opt.Models = tiers.Models{Routine: id}
	}
	return tiers.New(tier, opt)
}

func printSuiteReport(out *os.File, rep suite.Report) {
	say(out, "--- boteval suite run -------------------------------------\n")
	say(out, "policy             %s\n", rep.Policy)
	say(out, "positions          %d (%d labelled, %d skipped)\n", rep.Positions, rep.Labelled, rep.Positions-rep.Labelled)
	say(out, "agreement          %d/%d (%.0f%%)\n", rep.Agree, rep.Labelled, rep.AgreeRate()*100)
	say(out, "reject-hits        %d\n", rep.RejectHits)
	say(out, "disagree           %d\n", rep.Disagree)
	say(out, "decline            %d\n", rep.Declines)
	say(out, "malformed          %d\n", rep.Malformed)
	say(out, "out-of-range       %d\n", rep.OutOfRange)
	say(out, "timeouts           %d\n", rep.Timeouts)
	say(out, "errors             %d\n", rep.Errors)
	say(out, "latency            p50 %v p99 %v max %v\n",
		rep.Latency.P50.Round(time.Microsecond), rep.Latency.P99.Round(time.Microsecond), rep.Latency.Max.Round(time.Microsecond))
	if rep.PromptBytesP50 > 0 {
		say(out, "prompt bytes p50   %d\n", rep.PromptBytesP50)
	}
	if rep.Tokens.InputTokens > 0 || rep.Tokens.OutputTokens > 0 {
		say(out, "tokens             in %d / out %d\n", rep.Tokens.InputTokens, rep.Tokens.OutputTokens)
	}
	tags := make([]string, 0, len(rep.ByTag))
	for t := range rep.ByTag {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	for _, t := range tags {
		s := rep.ByTag[t]
		say(out, "  tag %-16s %d/%d (%.0f%%)\n", t, s.Agree, s.Labelled, s.AgreeRate()*100)
	}
	for _, res := range rep.Results {
		if res.Outcome.Miss() {
			say(out, "  MISS %-28s %-12s chose %d %q; wanted %v\n", res.ID, res.Outcome, res.Index, res.Label, res.Want)
		}
	}
	say(out, "-----------------------------------------------------------\n")
}

// --- suite harvest ---------------------------------------------------

type suiteHarvestOpts struct {
	From   []string
	To     string
	Filter suite.HarvestFilter
}

func parseSuiteHarvest(args []string) (suiteHarvestOpts, error) {
	var (
		o                                    suiteHarvestOpts
		from, fallbacks, layers, seats, tags string
	)
	fs := flag.NewFlagSet("suite harvest", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&from, "from", "", "comma-separated decision logs or globs (also taken from positional arguments)")
	fs.StringVar(&o.To, "to", "", "directory to write unlabelled positions into (required)")
	fs.BoolVar(&o.Filter.Escalated, "escalated", false, "only windows that left Layer A")
	fs.BoolVar(&o.Filter.Disagree, "disagree", false, "only windows where the model and the heuristic wanted different moves")
	fs.StringVar(&fallbacks, "fallback", "", "comma-separated fallback causes: malformed,out-of-range,timeout,no-budget,error")
	fs.StringVar(&layers, "layer", "", "comma-separated layers: A,B,C,random")
	fs.StringVar(&seats, "seat", "", "comma-separated seat indices")
	fs.StringVar(&tags, "tag", "", "comma-separated auto-tags: block,attack,choice,land,mulligan,cast,stack-response")
	fs.IntVar(&o.Filter.Limit, "limit", 0, "write at most N positions (sampled deterministically with --seed)")
	fs.Int64Var(&o.Filter.Seed, "seed", 1, "sampling seed")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	o.From = append(splitList(from), fs.Args()...)
	if len(o.From) == 0 {
		return o, fmt.Errorf("--from needs at least one decision log")
	}
	if strings.TrimSpace(o.To) == "" {
		return o, fmt.Errorf("--to needs an output directory")
	}
	o.Filter.Fallbacks = splitList(fallbacks)
	o.Filter.Layers = splitList(layers)
	o.Filter.Tags = splitList(tags)
	for _, s := range splitList(seats) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return o, fmt.Errorf("--seat %q is not a seat index", s)
		}
		o.Filter.Seats = append(o.Filter.Seats, n)
	}
	return o, nil
}

func runSuiteHarvest(args []string) int {
	o, err := parseSuiteHarvest(args)
	if err != nil {
		if err != flag.ErrHelp {
			say(os.Stderr, "boteval suite harvest: %v\n", err)
		}
		return 2
	}
	paths, err := expandLogs(o.From)
	if err != nil {
		say(os.Stderr, "%v\n", err)
		return 1
	}
	if len(paths) == 0 {
		say(os.Stderr, "no decision logs matched %v\n", o.From)
		return 1
	}
	rep, err := suite.Harvest(paths, o.Filter, o.To)
	if err != nil {
		say(os.Stderr, "%v\n", err)
		return 1
	}
	out := os.Stdout
	say(out, "--- boteval suite harvest ---------------------------------\n")
	say(out, "logs               %d\n", len(rep.Files))
	say(out, "records scanned    %d (%d carried a full view)\n", rep.Records, rep.Full)
	say(out, "candidates         %d\n", rep.Candidates)
	say(out, "written            %d → %s\n", rep.Written, o.To)
	say(out, "-----------------------------------------------------------\n")
	say(out, "Every position is UNLABELLED. Render one, decide what the right\nmove is, and fill in `expected.accept`:\n")
	say(out, "  boteval suite render --pos %s/<id>.json\n", strings.TrimRight(o.To, "/"))
	return 0
}

// expandLogs turns globs and directories into a sorted, deduped list
// of decision-log files.
func expandLogs(patterns []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, pat := range patterns {
		info, err := os.Stat(pat)
		switch {
		case err == nil && info.IsDir():
			matches, gerr := filepath.Glob(filepath.Join(pat, "*.decisions.jsonl*"))
			if gerr != nil {
				return nil, fmt.Errorf("glob %s: %w", pat, gerr)
			}
			for _, m := range matches {
				add(m)
			}
		case err == nil:
			add(pat)
		default:
			matches, gerr := filepath.Glob(pat)
			if gerr != nil {
				return nil, fmt.Errorf("glob %s: %w", pat, gerr)
			}
			for _, m := range matches {
				add(m)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// --- suite render ----------------------------------------------------

type suiteRenderOpts struct {
	Pos  string
	Deck string
	Dump string
}

func parseSuiteRender(args []string) (suiteRenderOpts, error) {
	var o suiteRenderOpts
	fs := flag.NewFlagSet("suite render", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&o.Pos, "pos", "", "position file to render (required)")
	fs.StringVar(&o.Deck, "deck", "izzet-aggro", "curated deck whose list becomes the static prompt block")
	fs.StringVar(&o.Dump, "dump", "", "Scryfall bulk dump for oracle text (default: $CMDCTRL_SCRYFALL_DUMP)")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if strings.TrimSpace(o.Pos) == "" {
		if fs.NArg() == 1 {
			o.Pos = fs.Arg(0)
		} else {
			return o, fmt.Errorf("--pos needs a position file")
		}
	}
	return o, nil
}

func runSuiteRender(args []string) int {
	o, err := parseSuiteRender(args)
	if err != nil {
		if err != flag.ErrHelp {
			say(os.Stderr, "boteval suite render: %v\n", err)
		}
		return 2
	}
	p, err := suite.LoadFile(o.Pos)
	if err != nil {
		say(os.Stderr, "%v\n", err)
		return 1
	}
	out := os.Stdout
	idx := loadIndex(out, o.Dump)
	profile, _ := buildProfile(idx, o.Deck)
	cfg := model.DefaultConfig()
	cfg.Deck = profile
	if err := suite.Render(out, p, cfg); err != nil {
		say(os.Stderr, "render: %v\n", err)
		return 1
	}
	return 0
}

func envDuration(name string) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(os.Getenv(name)))
	if err != nil {
		return 0
	}
	return d
}
