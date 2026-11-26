package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turntimber Grove — Land (Duel Decks: Zendikar vs. Eldrazi, #5548):
//
//	"This land enters tapped.
//	 When this land enters, target creature gets +1/+1 until end of turn.
//	 {T}: Add {G}."
//
// The trigger has a target chosen as it goes on the stack, and is removed
// from the stack if there is no creature to choose (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Turntimber Grove — target creature gets +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "62934aab-b6fa-41b2-ac05-d3fa0e3c5a79",
		Name:         "Turntimber Grove",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: label}.Apply(ctx)
			}), effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("G")},
	})
}
