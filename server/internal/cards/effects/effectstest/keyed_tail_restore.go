package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_tail_restore_test.go — #4253 slice 3: a CR 616 prompt whose
// event carries a continuation is a restore point once that
// continuation is data (game/keyed_tail.go). Each test takes a STRICT
// restore (RestoreStrict, the boot path) while the prompt is open,
// answers it on the restored table, and checks what a player sees
// afterwards.

// searchEventsSince counts EventSearchLibrary from `from` on, and
// reports whether the last one says the library was shuffled.
func SearchEventsSince(g *game.Game, from int) (n int, shuffled bool) {
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary {
			n++
			shuffled = ev.Label == game.SearchShuffledLabel
		}
	}
	return n, shuffled
}

// A search whose "then" is a registered tail keeps it across the
// restore: the tail is told the cards that entered.
var KeyedSearchThenProbe = game.RegisterTail("test/search-then-probe", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return g.AddCounterForEffect(o.IDs[0], game.CounterPlusOne, o.Tail.N)
})

// libraryShuffles counts the plain shuffles (ShuffleLibraryForEffect).
func LibraryShuffles(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Label == "shuffle" {
			n++
		}
	}
	return n
}

// keyedTokenThenProbe taps every token it is told about, so a test can
// read that the creation's keyed continuation ran with the tokens.
var KeyedTokenThenProbe = game.RegisterTail("test/token-then-probe", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.TapTargetForEffect(id); err != nil {
			return err
		}
	}
	return nil
})

// keyedLifeThenProbe puts as many +1/+1 counters on Carry[0] as the life
// change moved, so a test can read the amount the tail was told.
var KeyedLifeThenProbe = game.RegisterTail("test/life-then-probe", func(g *game.Game, o game.TailOutcome) error {
	return g.AddCounterForEffect(o.Tail.Carry[0], game.CounterPlusOne, o.Amount)
})

// lifeLossPlusOneProbeOracle is a test card: "If an opponent would lose
// life, they lose that much life plus 1 instead." Beside Bloodletter of
// Aclazotz it makes each opponent's loss during your turn an ordering
// prompt, which no two printed cards in the catalog do yet.
const LifeLossPlusOneProbeOracle = "test-4253-life-loss-plus-one"

func init() {
	effects.Register(effects.Spec{
		OracleID: LifeLossPlusOneProbeOracle,
		Name:     "Life Loss Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeDelta < 0 && src != nil && ev.LifePlayer != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.LifeDelta--
				return nil
			},
			Label: "Life Loss Probe: lose one more",
		}},
	})
}

// --- a face-down exile batch (#4257, #4261) ----------------------------------

// exileOrderProbeOracles are two test cards, each "if a card named Needle
// would be put into exile, it is put into exile instead": two different
// declared effects, so a Needle's exile asks for their order (CR 616.1).
// They are catalog rows, so a restored table gathers them again.
var exileOrderProbeOracles = []string{"test-4253-exile-order-a", "test-4253-exile-order-b"}

func init() {
	for i, oracle := range exileOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Exile Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneExile {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Exile Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// seedExileOrderProbes puts both probes on the battlefield under `who`.
func SeedExileOrderProbes(g *game.Game, who uuid.UUID) {
	for i, oracle := range exileOrderProbeOracles {
		SeedReplacementPermanent(g, oracle, "Exile Order Probe "+string(rune('A'+i)), who)
	}
}

// --- a keyword action with a "then" ------------------------------------------

// strictRestoreWithScryThen takes a strict restore and checks the open
// look-at was written with its keyed "then".
func StrictRestoreWithScryThen(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if (c.Kind == game.PendingChoiceScry || c.Kind == game.PendingChoiceSurveil) && c.PickThen != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no look-at in the snapshot carries a keyed then: %+v", snap.Continuations)
	}
	return RestoreRoundTrip(t, g, true)
}
