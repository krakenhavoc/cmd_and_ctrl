package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helium Squirter — Creature — Beast Mutant {4}{U}, 0/0 (Ravnica: Clue
// Edition, #5929):
//
//	"Graft 3 (This creature enters with three +1/+1 counters on it.
//	 Whenever another creature enters, you may move a +1/+1 counter from
//	 this creature onto it.)
//	 {1}: Target creature with a +1/+1 counter on it gains flying until
//	 end of turn."
//
// Graft is Cytoplast Manipulator's entry replacement plus cmdr2B23Graft
// (CR 702.58a). The target must have a +1/+1 counter as it is chosen and
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a65ba9bd-1bc6-41c1-a2a4-04b97a5e9541",
		Name:         "Helium Squirter",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Graft 3 — enters with three +1/+1 counters on it."),
		},
		Triggered: []game.TriggeredAbility{effects.Cmdr2B23Graft("Helium Squirter")},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}: Target creature with a +1/+1 counter on it gains flying until end of turn.",
			Purpose: game.Purpose{Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Keywords: []string{"flying"}},
			})},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature with a +1/+1 counter on it", effects.Not(effects.B13WithoutPlusOneCounter())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Helium Squirter — flying"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
