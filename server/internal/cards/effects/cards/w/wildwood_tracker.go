package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildwood Tracker — Creature — Elf Warrior {G}, 1/1 (Throne of Eldraine,
// #5516):
//
//	"Whenever this creature attacks or blocks, if you control another
//	 non-Human creature, this creature gets +1/+1 until end of turn."
//
// An intervening "if" (CR 603.4): checked as it would trigger and again as it
// resolves. "Another" is object identity, never a name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ca31946-72ef-4c45-a7bb-494a99dc2982",
		Name:         "Wildwood Tracker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(
				effects.CondBoard{Name: "you control another non-Human creature", Test: func(ctx *effects.Context) bool {
					word := ctx.TextSubtype("Human")
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == ctx.Controller() && c.InstanceID != ctx.Source() && c.IsCreature() && !c.HasSubtype(word) {
							return true
						}
					}
					return false
				}},
				effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
					"Wildwood Tracker — this creature gets +1/+1 until end of turn",
					effects.ThisGetsUntilEndOfTurn(1, 1, "Wildwood Tracker — +1/+1"))),
		},
	})
}
