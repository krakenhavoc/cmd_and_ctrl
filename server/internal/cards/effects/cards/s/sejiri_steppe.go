package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sejiri Steppe — Land (Worldwake, S82 OLD batch 490, #5582):
//
//	"This land enters tapped.
//	 When this land enters, target creature you control gains protection from
//	 the color of your choice until end of turn.
//	 {T}: Add {W}."
//
// The tapped entry is the CR 614 self-replacement (SelfEntersTapped); the
// trigger targets a creature you control and asks for the colour as it
// resolves (ProtectionFromAChosenColorForTheTarget, as Mother of Runes does).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dfbf95e-a91b-429c-96e2-95ac777e7027",
		Name:         "Sejiri Steppe",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sejiri Steppe — target creature you control gains protection from the color of your choice until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ProtectionFromAChosenColorForTheTarget("Sejiri Steppe")(item, effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
