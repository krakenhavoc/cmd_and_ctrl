package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soaring Seacliff — Land (Commander 2020, S82 OLD, #5513):
//
//	"This land enters tapped.
//	 When this land enters, target creature gains flying until end of turn.
//	 {T}: Add {U}."
//
// Smoldering Spires' tapped land with an entry trigger: the target is chosen as
// the trigger goes on the stack and gains flying (a layer 6 grant) until the
// end of the turn if it is still legal as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Soaring Seacliff — target creature gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "a37544b6-0048-4213-8e40-76ba8a0b6d1b",
		Name:         "Soaring Seacliff",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: label}.Apply(ctx)
			}), effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("U")},
	})
}
