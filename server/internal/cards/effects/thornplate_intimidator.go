package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornplate Intimidator — Creature — Rat Rogue {3}{B}, 4/3 (Bloomburrow,
// #4467):
//
//	"Offspring {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, when this creature enters, create a 1/1 token copy of it.)
//	 When this creature enters, target opponent loses 3 life unless they
//	 sacrifice a nonland permanent of their choice or discard a card."
//
// Offspring is the shared keyword (offspring.go), and the token copy
// brings its own entry trigger, so a paid offspring asks twice. The entry
// trigger is Torment of Hailfire's three-way punisher asked once, of the
// targeted opponent (punisherRepeat.onTargetPlayer): losing 3 life is
// always offered, the sacrifice only with a nonland permanent to give and
// the discard only with a card in hand (CR 608.2). An opponent who has
// left by resolution is asked nothing (CR 608.2b).
//
// No simplification.
func init() {
	registerPunisher(thornplateIntimidatorPunisher, "thornplate-intimidator")
	Register(Spec{
		OracleID:      "33b6555e-035d-4241-a341-a9b9139ca6f0",
		Name:          "Thornplate Intimidator",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{3}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Thornplate Intimidator"),
			Targeting(WhenThisEnters("Thornplate Intimidator — target opponent loses 3 life unless they sacrifice a nonland permanent or discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return thornplateIntimidatorPunisher.onTargetPlayer(NewContext(g, item))
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// thornplateIntimidatorPunisher is the printed punisher: 3 life a refusal.
var thornplateIntimidatorPunisher = punisherRepeat{Name: "Thornplate Intimidator", Life: 3}
