package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brion Stoutarm — Legendary Creature — Giant Warrior {2}{R}{W}, 4/4 (Secret
// Lair Drop, #5462):
//
//	"Lifelink
//	 {R}, {T}, Sacrifice another creature: Brion Stoutarm deals damage equal
//	 to the sacrificed creature's power to target player or planeswalker."
//
// Bloodshot Cyclops' fling: the amount is the sacrificed creature's power as
// it last existed on the battlefield, floored at zero (CR 107.1b), over a
// target restricted to a player or a planeswalker. {T} needs Brion to have been
// under its controller's control since their most recent turn began
// (CR 302.6), and the sacrifice can't be Brion himself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b816b3cc-ae4b-4fb7-8b1a-e01ab83459a3",
		Name:            "Brion Stoutarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice another creature: Brion Stoutarm deals damage equal to the sacrificed creature's power to target player or planeswalker.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEqualToSacrificedPower(item, effects.NewContext(g, item))
			},
		}},
	})
}
