package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// arena.go is `boteval arena`: N headless bot-vs-bot games and the
// report block ADR 0052 asks every Part 2 PR to carry.
//
// The subcommand's job is resolution and artifacts — turning flags
// and environment into a botarena.Config, and turning the results
// into files somebody can read next week. The playing is all in
// internal/botarena.
//
// Artifacts go under <out>/<RFC3339 start>/, one directory per run,
// so two runs never overwrite each other and a directory listing is a
// history:
//
//	summary.json   the whole Summary, machine-readable
//	summary.md     the report block, for pasting into a PR
//	games.jsonl    one full GameResult per line, WRITTEN AS EACH GAME
//	               ENDS (a ten-game model run is an hour; a harness
//	               that writes nothing until the end is one a Ctrl-C
//	               destroys). OPERATOR-ONLY: a stalled game carries a
//	               dump of every seat's legal moves, so the file names
//	               castable cards in all four hands.
//	decisions/     per-game decision logs, when --decision-log is on
//	replays/       per-game replay JSONL, when --replays is on — and
//	               games/ and restore/ beside it, which ws.Room writes
//	               off the same root
//
// The run directory is 0700 and games.jsonl is 0600, the modes ADR
// 0052 gives the decision log, because they hold the same class of
// aggregated hidden information.

// localDefaultMaxThink mirrors cmd/server/main.go. A self-hosted
// model cannot answer a Commander window in the tier's 2s default, so
// a run against a local endpoint that did not say otherwise would
// measure nothing but timeouts — every window falling back to the
// heuristic under the model tier's name.
const localDefaultMaxThink = 20 * time.Second

// arenaFlags is the parsed command line. It is its own type, and
// parsing is its own function, so that the resolution rules (env
// fallbacks, the local think default, deck/seat arity) are testable
// without a model endpoint or a 600 MiB dump.
type arenaFlags struct {
	seats    []tiers.Tier
	decks    []string
	names    []string
	games    int
	seed     uint64
	rotate   bool
	turns    int
	wall     time.Duration
	stall    time.Duration
	maxThink time.Duration
	// blockGrace is the declare-blockers hold. 0 or less is OFF; see
	// config, where it has to be spelled negative for the arena.
	blockGrace time.Duration
	modelID    string
	frontier   string
	endpoint   string
	out        string
	decLog     bool
	decMode    string
	replays    bool
	dump       string
	note       string
	printMD    bool
	printJSON  bool

	// needsModel and needsIndex are conclusions parse drew, kept so
	// the caller does not re-derive them.
	needsModel bool
	needsIndex bool
	// thinkDefaulted records that the local 20s default was applied,
	// so the run can say so out loud exactly once.
	thinkDefaulted bool
}

func parseArenaFlags(args []string, out io.Writer) (*arenaFlags, error) {
	fs := flag.NewFlagSet("arena", flag.ContinueOnError)
	fs.SetOutput(out)
	seats := fs.String("seats", "", "comma-separated tiers, one per chair (e.g. assisted,heuristic,heuristic,heuristic)")
	decks := fs.String("decks", "", "comma-separated curated deck ids, one per chair; empty deals the synthetic battle deck")
	names := fs.String("names", "", "comma-separated tally names, one per chair; empty tallies each chair under its tier")
	games := fs.Int("games", 10, "how many games to play")
	seed := fs.Uint64("seed", 1, "seed of the first game; game i uses seed+i")
	rotate := fs.Bool("rotate", false, "move each contestant one chair along per game, so turn order cancels")
	turns := fs.Int("turn-budget", 60, "stop a game that has not ended by this turn")
	wall := fs.Duration("wall", 30*time.Minute, "per-game wall clock")
	stall := fs.Duration("stall", 0, "declare a stall after this long with no committed move (default 3×max-think+15s)")
	maxThink := fs.Duration("max-think", 0, "model tiers' per-window deadline (default: $CMDCTRL_BOT_MAX_THINK, or 20s against a local endpoint)")
	modelID := fs.String("model", "", "routine model id (default: $CMDCTRL_BOT_MODEL)")
	frontier := fs.String("frontier-model", "", "escalated model id (default: $CMDCTRL_BOT_FRONTIER_MODEL, else --model)")
	endpoint := fs.String("endpoint", "", "OpenAI-compatible endpoint (default: $CMDCTRL_OPENAI_ENDPOINT)")
	out2 := fs.String("out", "", "directory for the run's artifacts; empty prints the report and writes nothing")
	decLog := fs.Bool("decision-log", false, "write a per-game decision log under <out>/decisions (operator-only: see docs/bot.md)")
	decMode := fs.String("decision-log-mode", "", "escalated (default) | all | model")
	replays := fs.Bool("replays", false, "write per-game replays under <out>: replays/<id>.jsonl (~320 MiB per four-seat game) plus games/<id>.json, a full authoritative-state marshal rewritten on every committed move, and restore/<id>.json while a game is live")
	blockGrace := fs.Duration("block-grace", aiseat.DefaultConfig().BlockGrace, "how long an attacking bot holds its pass in declare-blockers while a defender is still declaring blockers; 0 or less turns it off, which is faster and (since #1279) loses no blocks")
	dump := fs.String("dump", "", "Scryfall bulk dump, needed by curated decks (default: $CMDCTRL_SCRYFALL_DUMP)")
	note := fs.String("note", "", "free-form note recorded in the report (model quantisation, what is being tested)")
	printMD := fs.Bool("md", false, "print the Markdown report to stdout (the default when neither --md nor --json is given)")
	printJSON := fs.Bool("json", false, "print the summary as JSON to stdout")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	a := &arenaFlags{
		games: *games, seed: *seed, rotate: *rotate, turns: *turns, wall: *wall,
		stall: *stall, maxThink: *maxThink, blockGrace: *blockGrace, out: *out2,
		decLog: *decLog, decMode: *decMode,
		replays: *replays, note: *note, printMD: *printMD, printJSON: *printJSON,
	}
	var err error
	if a.seats, err = parseSeats(*seats); err != nil {
		return nil, err
	}
	if a.decks, err = parseDecks(*decks, len(a.seats)); err != nil {
		return nil, err
	}
	if a.names, err = parseNames(*names, len(a.seats)); err != nil {
		return nil, err
	}
	for _, t := range a.seats {
		if t.NeedsModel() {
			a.needsModel = true
		}
	}
	for _, d := range a.decks {
		if d != "" {
			a.needsIndex = true
		}
	}
	// A model seat wants the deck PROFILE too, which is built from
	// the same index; without it the prompt carries card names only.
	a.needsIndex = a.needsIndex || (a.needsModel && anyDeck(a.decks))

	a.endpoint = firstNonEmpty(*endpoint, os.Getenv(model.EnvOpenAIEndpoint))
	a.modelID = firstNonEmpty(*modelID, os.Getenv("CMDCTRL_BOT_MODEL"))
	a.frontier = firstNonEmpty(*frontier, os.Getenv("CMDCTRL_BOT_FRONTIER_MODEL"))
	a.dump = firstNonEmpty(*dump, os.Getenv("CMDCTRL_SCRYFALL_DUMP"))
	if a.maxThink == 0 {
		if d, derr := time.ParseDuration(strings.TrimSpace(os.Getenv("CMDCTRL_BOT_MAX_THINK"))); derr == nil && d > 0 {
			a.maxThink = d
		}
	}
	if a.endpoint != "" && a.maxThink == 0 {
		a.maxThink, a.thinkDefaulted = localDefaultMaxThink, true
	}
	if _, merr := decisionlog.ParseMode(a.decMode); merr != nil {
		return nil, merr
	}
	if a.games < 1 {
		return nil, fmt.Errorf("--games must be at least 1, got %d", a.games)
	}
	if (a.decLog || a.replays) && a.out == "" {
		return nil, errors.New("--decision-log and --replays need somewhere to write: pass --out DIR")
	}
	if a.printMD && a.printJSON {
		return nil, errors.New("--md and --json are mutually exclusive")
	}
	if !a.printMD && !a.printJSON {
		a.printMD = true
	}
	return a, nil
}

func parseSeats(s string) ([]tiers.Tier, error) {
	fields := splitList(s)
	if len(fields) == 0 {
		return nil, errors.New("--seats is required (e.g. --seats heuristic,heuristic)")
	}
	out := make([]tiers.Tier, 0, len(fields))
	for _, f := range fields {
		t, err := tiers.Parse(f)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// parseDecks returns one deck id per seat. An empty list deals the
// synthetic battle deck to everyone, which is what the ungated tests
// and a quick smoke run want. A partial list is refused rather than
// padded: a run where two seats hold curated decks and two hold
// vanilla bears is not a measurement of anything, and silently
// producing one is worse than an error message.
func parseDecks(s string, seats int) ([]string, error) {
	fields := splitList(s)
	if len(fields) == 0 {
		return make([]string, seats), nil
	}
	if len(fields) != seats {
		return nil, fmt.Errorf("--decks has %d entries for %d seats; give one per seat or none at all", len(fields), seats)
	}
	return fields, nil
}

// parseNames returns one tally name per seat, or all-empty for "tally
// each chair under its tier".
//
// It exists for the run where every chair is the SAME tier and the
// thing being compared is something else — four heuristic seats on
// four curated decks, or two configurations of `assisted`. Without it
// that run collapses into a single row that says "heuristic won 25%
// of a four-seat table", which is arithmetic rather than a finding.
func parseNames(s string, seats int) ([]string, error) {
	fields := splitList(s)
	if len(fields) == 0 {
		return make([]string, seats), nil
	}
	if len(fields) != seats {
		return nil, fmt.Errorf("--names has %d entries for %d seats; give one per seat or none at all", len(fields), seats)
	}
	return fields, nil
}

func splitList(s string) []string {
	out := []string{}
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func anyDeck(d []string) bool {
	for _, s := range d {
		if s != "" {
			return true
		}
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

// config turns parsed flags into the arena's configuration. idx and
// client are passed in rather than resolved here so that the
// expensive halves (a 600 MiB dump, a live endpoint) stay out of the
// tests.
func (a *arenaFlags) config(idx *cards.Index, client model.Client, dl *decisionlog.Logger, runDir string, log *slog.Logger) botarena.Config {
	seats := make([]botarena.SeatSpec, 0, len(a.seats))
	for i, t := range a.seats {
		seats = append(seats, botarena.SeatSpec{Tier: t, Deck: a.decks[i], Name: a.names[i]})
	}
	cfg := botarena.Config{
		Seats: seats, Games: a.games, Seed: a.seed, Rotate: a.rotate,
		TurnBudget: a.turns, Wall: a.wall, Stall: a.stall,
		Index: idx, Client: client, MaxThink: a.maxThink,
		Models:      tiers.Models{Routine: a.modelID, Frontier: a.frontier},
		DecisionLog: dl, Log: log,
	}
	// botarena.Config.Runner's zero BlockGrace means "production's
	// 4s", so an operator asking for none has to be spelled with a
	// negative duration — otherwise `--block-grace 0` would turn
	// itself back on.
	cfg.Runner.BlockGrace = a.blockGrace
	if a.blockGrace <= 0 {
		cfg.Runner.BlockGrace = -1
	}
	if a.replays && runDir != "" {
		cfg.ReplayDir = runDir
	}
	return cfg
}

func runArena(args []string) int {
	out := os.Stdout
	a, err := parseArenaFlags(args, os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "boteval arena: %v\n", err)
		return 2
	}
	progress := io.Writer(out)
	if a.printJSON {
		progress = os.Stderr
	}

	// A model tier with no endpoint is refused rather than quietly
	// downgraded — the same rule the lobby applies to a tier whose
	// model is unavailable. An `assisted` seat with a nil client
	// plays Layer A + B and reports itself as `assisted`, so the run
	// would produce a heuristic's win rate under a model's name.
	var client model.Client
	if a.needsModel {
		// The second return is the chat-completions URL the transport
		// resolved, not an error — it is reported rather than dropped
		// so that "no endpoint is configured" is never printed at an
		// operator who did set one, and so a typo'd endpoint shows up
		// here rather than as a per-window warning an hour in.
		c, url := buildClient(a.endpoint)
		if c == nil || url == "" {
			if a.endpoint != "" {
				fmt.Fprintf(os.Stderr, "boteval arena: --endpoint %q did not resolve to a chat-completions URL\n", a.endpoint)
				return 2
			}
			fmt.Fprintf(os.Stderr, "boteval arena: a model tier was asked for but no endpoint is configured: pass --endpoint or set %s\n", model.EnvOpenAIEndpoint)
			return 2
		}
		say(progress, "model endpoint: %s\n", url)
		if a.modelID == "" {
			fmt.Fprintln(os.Stderr, "boteval arena: a model tier was asked for but no model id is configured: pass --model or set CMDCTRL_BOT_MODEL")
			return 2
		}
		client = c
		if a.thinkDefaulted {
			say(progress, "bot think deadline raised to %s for the local model transport; pass --max-think to choose your own\n", a.maxThink)
		}
		if a.maxThink < 5*time.Second {
			say(progress, "WARNING: --max-think %s is short for a self-hosted model; windows that overrun it play the HEURISTIC's move under the model tier's name\n", a.maxThink)
		}
	}

	var idx *cards.Index
	if a.needsIndex {
		if idx = loadIndex(progress, a.dump); idx == nil && anyDeck(a.decks) {
			fmt.Fprintln(os.Stderr, "boteval arena: curated decks need a Scryfall dump: pass --dump or set CMDCTRL_SCRYFALL_DUMP")
			return 2
		}
	}

	started := time.Now().UTC()
	runDir := ""
	if a.out != "" {
		if runDir, err = createArenaRunDir(a.out, started); err != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", err)
			return 1
		}
	}

	var dl *decisionlog.Logger
	if a.decLog {
		dl, err = decisionlog.New(decisionlog.Options{Dir: filepath.Join(runDir, "decisions"), Mode: decisionlog.Mode(a.decMode)})
		if err != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", err)
			return 1
		}
		say(progress, "DECISION LOG IS ON: %s — operator-only; it aggregates every seat's own view of a game and must never be attached to a bug report.\n", dl.Dir())
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cfg := a.config(idx, client, dl, runDir, log)
	cfg.Note = a.note
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "boteval arena: %v\n", err)
		return 2
	}

	// Ctrl-C stops between games rather than killing the process, so
	// an interrupted run still writes the summary of what it played.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var jsonl *os.File
	if runDir != "" {
		if jsonl, err = createGamesJSONL(runDir); err != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", err)
			return 1
		}
		defer func() { _ = jsonl.Close() }()
		say(progress, "games.jsonl IS OPERATOR-ONLY: a stalled game's entry dumps every seat's legal moves, so the file names castable cards in every hand at the table; never attach it to a bug report.\n")
	}

	say(progress, "arena: %d games, seats [%s], seed %d, rotation %s\n",
		a.games, strings.Join(tierNames(a.seats), ", "), a.seed, onOff(a.rotate))
	// The operator's --games is never changed for them; the run says
	// what the seating will actually be, and the report keeps the
	// histogram.
	if w := botarena.ChairBalanceWarning(len(a.seats), a.games, a.rotate); w != "" {
		say(progress, "WARNING: %s\n", w)
	}

	enc := (*json.Encoder)(nil)
	if jsonl != nil {
		enc = json.NewEncoder(jsonl)
	}
	played := 0
	sum, runErr := botarena.Run(ctx, cfg, func(r botarena.GameResult) {
		played++
		verdict := "no single survivor"
		if w := r.WinnerLabel(); w != "" {
			verdict = w + " won"
		}
		if r.Stalled {
			verdict = "STALLED"
		}
		say(progress, "  game %d/%d seed %d: %s at turn %d in %s\n",
			played, a.games, r.Seed, verdict, r.Turns, r.Elapsed.Round(time.Millisecond))
		if enc != nil {
			if eerr := enc.Encode(r); eerr != nil {
				fmt.Fprintf(os.Stderr, "boteval arena: writing games.jsonl: %v\n", eerr)
			}
			_ = jsonl.Sync()
		}
	})
	exit := arenaExit(runErr)
	if exit != 0 {
		fmt.Fprintf(os.Stderr, "boteval arena: %v\n", runErr)
	}

	md := sum.Markdown()
	if runDir != "" {
		if werr := os.WriteFile(filepath.Join(runDir, "summary.md"), []byte(md), 0o644); werr != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", werr)
		}
		blob, merr := json.MarshalIndent(sum, "", "  ")
		if merr == nil {
			merr = os.WriteFile(filepath.Join(runDir, "summary.json"), blob, 0o644)
		}
		if merr != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", merr)
		}
	}
	if a.printMD {
		say(out, "\n%s", md)
	}
	if a.printJSON {
		blob, merr := json.MarshalIndent(sum, "", "  ")
		if merr != nil {
			fmt.Fprintf(os.Stderr, "boteval arena: %v\n", merr)
		} else {
			say(out, "%s\n", blob)
		}
	}
	if runDir != "" {
		say(progress, "\nartifacts: %s\n", runDir)
	}
	return exit
}

// arenaExit is the process's answer to "did this run do what it was
// asked to".
//
// A Ctrl-C is the operator's own decision and is not a failure: the
// run stops between games and the summary describes what it played.
// Anything else is, EVEN WHEN games were played — a run whose sixth
// game could not be started measured six-tenths of what was asked
// for, and a script that reads exit 0 off it will file the report as
// if it were whole. The summary is still written either way: the
// games that played are the evidence.
func arenaExit(runErr error) int {
	if runErr == nil || errors.Is(runErr, context.Canceled) {
		return 0
	}
	return 1
}

// createGamesJSONL opens the run's games.jsonl at 0600.
//
// A stalled game's GameResult carries a StallDump that enumerates
// every seat's legal moves, and the cast moves are enumerated from
// each hand — so the file names castable cards in all four hands.
// That is the same aggregated-hidden-information class ADR 0052
// protects the decision log with, and this is the file an operator
// would reach for when reporting a stall.
func createGamesJSONL(dir string) (*os.File, error) {
	return os.OpenFile(filepath.Join(dir, "games.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}

func createArenaRunDir(root string, started time.Time) (string, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	base := started.UTC().Format(time.RFC3339Nano)
	for n := 1; ; n++ {
		name := base
		if n > 1 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		dir := filepath.Join(root, name)
		// 0700: the run directory holds games.jsonl (stall dumps that
		// name every seat's hand), and with --decision-log or
		// --replays it holds whole-table state as well. Same mode ADR
		// 0052 gives the decision log, for the same reason.
		if err := os.Mkdir(dir, 0o700); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
}

func tierNames(t []tiers.Tier) []string {
	out := make([]string, 0, len(t))
	for _, v := range t {
		out = append(out, string(v))
	}
	return out
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
