package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ozolith, the Shattered Spire — Legendary Artifact {1}{G} (EDHREC top
// 1,000, #3654):
//
//	"If one or more +1/+1 counters would be put on an artifact or
//	 creature you control, that many plus one +1/+1 counters are put on
//	 it instead.
//	 {1}{G}, {T}: Put a +1/+1 counter on target artifact or creature
//	 you control. Activate only as a sorcery."
//
// Hardened Scales widened from creatures to artifacts too: the same
// counter-placement replacement, with the type gate loosened. The
// activated ability is an ordinary sorcery-speed counter placement, so
// it goes through the same pipeline and gets its own extra counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41fed659-237c-4d8e-ad31-d17fa0d3f764",
		Name:         "Ozolith, the Shattered Spire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 || ev.CounterName != "+1/+1" {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && (target.IsCreature() || target.IsArtifact()) && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ozolith, the Shattered Spire: +1 +1/+1 counter",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{1}{G}, {T}: Put a +1/+1 counter on target artifact or creature you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{G}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
			Effect:       PlusOneCounterOnChosenTarget,
		}},
	})
}
