package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_counter_life_restore_test.go — #4253 slice 7: the "then" of a
// counter placement and of a life change is a registered tail, so the CR 616
// order prompt that two different counter (or life) replacements ask is a
// restore point. Each test takes a STRICT restore while the prompt is open,
// answers it on the restored table and checks the card still finishes off the
// disk. (strictRestoreWithFrame, answerOrderBySource and the replacement
// seeding helpers are in replacement_resume_restore_test.go,
// counter_tail_cards_test.go and doubling_season_test.go.)

// counterPlusOneProbeOracle is a test card: "If a counter would be put on a
// permanent you control, put one more counter on it instead." Beside Doubling
// Season it makes a placement on any permanent (an enchantment, a land) ask
// for the order, which Hardened Scales only does for a creature.
const CounterPlusOneProbeOracle = "test-4253s7-counter-plus-one"

func init() {
	effects.Register(effects.Spec{
		OracleID: CounterPlusOneProbeOracle,
		Name:     "Counter Plus One Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Counter Plus One Probe: one more counter",
		}},
	})
}

// counterOrderBoard puts Hardened Scales and Doubling Season under `who`: a
// +1/+1 counter placement on one of their creatures asks for the order.
func CounterOrderBoard(g *game.Game, who uuid.UUID) (hs, ds uuid.UUID) {
	hs = SeedReplacementPermanent(g, HardenedScalesOracle, "Hardened Scales", who)
	ds = SeedReplacementPermanent(g, DoublingSeasonOracle, "Doubling Season", who)
	return hs, ds
}
