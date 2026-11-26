package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeering Instigator — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Goblin Rogue {1}{R}, 2/1: Morph {2}{R}. When this creature is turned
//	face up, if it's your turn, gain control of another target creature until
//	end of turn. Untap that creature. It gains haste until end of turn.
//
// CR 603.4: the "if it's your turn" is an intervening-if, checked when the
// trigger fires and again as it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "170adfe9-ffbe-40b1-b665-e756e7eb0302",
		Name:             "Jeering Instigator",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(jeeringInstigatorTrigger(), effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}

func jeeringInstigatorTrigger() game.TriggeredAbility {
	const label = "Jeering Instigator — gain control of another target creature until end of turn, untap it, haste"
	t := effects.WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
		if !effects.IsActivePlayer(g, item.Controller) {
			return nil
		}
		return effects.ThreatenTheTarget("Jeering Instigator")(item, effects.NewContext(g, item))
	})
	t.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.Self(ev, source, game.Characteristic{}, g) && effects.IsActivePlayer(g, source.Controller)
	}
	return t
}
