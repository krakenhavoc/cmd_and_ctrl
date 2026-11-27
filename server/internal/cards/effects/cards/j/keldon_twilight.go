package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Twilight — Enchantment {1}{B}{R} (Planeshift, S82 OLD b644, #5615):
//
//	"At the beginning of each player's end step, if no creatures attacked this
//	 turn, that player sacrifices a creature of their choice that they controlled
//	 since the beginning of the turn."
//
// The trigger fires at every player's end step with an intervening "if" (CR 603.4,
// OldB644IfNoCreaturesAttacked: it counts attackers that have since left), and the
// player whose step it is chooses their own victim among the creatures they have
// controlled continuously since the turn began. A player with none sacrifices
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c886a076-dbe2-4d80-8dab-537b84d0837a",
		Name:         "Keldon Twilight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB644IfNoCreaturesAttacked(effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Keldon Twilight — that player sacrifices a creature that they controlled since the beginning of the turn",
				effects.OldB644TriggerPlayerSacrifices("Keldon Twilight — sacrifice a creature you controlled since the beginning of the turn",
					func(c game.Card) bool { return c.IsCreature() && !c.SummonedThisTurn }))),
		},
	})
}
