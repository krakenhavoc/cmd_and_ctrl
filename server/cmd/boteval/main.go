// Command boteval is the bot evaluation harness: the tools that
// measure what the AI bot seat actually does, as opposed to what ADR
// 0033 says it should.
//
// It is a separate binary rather than a test because the interesting
// runs need things CI does not have — a model endpoint, a Scryfall
// dump, minutes of wall clock — and because their output is a report
// somebody reads, not a pass/fail.
//
//	boteval probe   one real request in the funnel's exact shape,
//	                against a configured endpoint, with the transport
//	                facts printed: prompt tokens vs the client-side
//	                estimate, finish_reason, whether a reasoning field
//	                came back, whether the reply parsed.
//
//	boteval arena   N headless bot-vs-bot games, with win rates and
//	                Wilson intervals, the funnel's own counters and
//	                the decision-latency tails, written out as a
//	                report block and a machine-readable summary.
//
// `suite` (a labelled position suite) is the remaining subcommand and
// lands in a later PR; the switch below is shaped for it.
//
// Build it with `make -C server build-boteval`.
//
// # Why cmd/ may do what internal/aiseat may not
//
// This package imports internal/game, which every policy package
// under aiseat/ is forbidden to (ADR 0033 §3, enforced by
// TestPolicyPackagesDoNotImportGame). That is not a loophole: the ban
// is on POLICIES, because a policy holding the authoritative state
// could read an opponent's hand. This binary is a harness — it builds
// a game, plays it, and hands each policy nothing but the filtered
// aiseat.Input the runner would.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch cmd := os.Args[1]; cmd {
	case "probe":
		os.Exit(runProbe(os.Args[2:]))
	case "arena":
		os.Exit(runArena(os.Args[2:]))
	case "suite":
		fmt.Fprintf(os.Stderr, "boteval: %q is not implemented yet (it lands with the arena and the position suite)\n", cmd)
		os.Exit(2)
	case "-h", "--help", "help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "boteval: unknown subcommand %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `boteval — the AI bot seat's evaluation harness

usage: boteval <subcommand> [flags]

subcommands:
  probe   send ONE request in the funnel's exact shape to the configured
          model endpoint and print what came back: prompt/completion
          tokens, finish_reason, whether a reasoning field was present,
          whether the reply parsed as an index, and wall time.

          boteval probe [--endpoint URL] [--model ID] [--max-tokens N]
                        [--deck izzet-aggro] [--dump path/to/default-cards.json]

          Env fallbacks: CMDCTRL_OPENAI_ENDPOINT, CMDCTRL_OPENAI_API_KEY,
          CMDCTRL_BOT_MODEL, CMDCTRL_SCRYFALL_DUMP.

  arena   play N bot-vs-bot games headlessly and report who won, how
          the funnel behaved and how long decisions took. Stalls are
          reported, not fatal.

          boteval arena --seats heuristic,heuristic [--decks a,b]
                        [--names a,b]
                        [--games N] [--seed N] [--rotate]
                        [--turn-budget N] [--wall 30m] [--stall 0]
                        [--max-think 20s] [--model ID] [--frontier-model ID]
                        [--endpoint URL] [--out DIR] [--decision-log]
                        [--decision-log-mode escalated|all|model]
                        [--replays] [--dump path] [--note text]
                        [--md] [--json]

          A model tier with no endpoint is REFUSED, not downgraded: a
          seat that quietly plays the heuristic under a model tier's
          name would corrupt the measurement rather than break it.

          Env fallbacks: CMDCTRL_OPENAI_ENDPOINT, CMDCTRL_OPENAI_API_KEY,
          CMDCTRL_BOT_MODEL, CMDCTRL_BOT_FRONTIER_MODEL,
          CMDCTRL_BOT_MAX_THINK, CMDCTRL_SCRYFALL_DUMP.

  suite   (not implemented yet)
`)
}
