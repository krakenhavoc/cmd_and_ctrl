package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defender en-Vec — Creature — Human Cleric {3}{W}, 2/4:
//
//	"Fading 4 (This creature enters with four fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 Remove a fade counter from this creature: Prevent the next 2 damage
//	 that would be dealt to any target this turn."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The ability's
// cost is the fade counter (RemoveAFadeCounterFromThis, paid as it is
// activated, CR 602.2b), so the counters it spends are ones the upkeep
// can no longer remove, and a Defender that spent them all is
// sacrificed at the next upkeep. The shield is Kitsune Healer's charged
// one (PreventNextDamage, CR 615.7), for 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb8a55a5-596e-427c-b745-4de8b0dfb095",
		Name:         "Defender en-Vec",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(4)},
		Triggered:    []game.TriggeredAbility{effects.FadingUpkeep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a fade counter from this creature: Prevent the next 2 damage that would be dealt to any target this turn.",
			Cost:    effects.RemoveAFadeCounterFromThis(),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.PreventNextDamage{Target: ts[0].ID, Amount: 2, Label: "Defender en-Vec: prevent the next 2 damage"}.Apply(ctx)
			},
		}},
	})
}
