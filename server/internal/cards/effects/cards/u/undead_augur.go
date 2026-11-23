package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Augur — Creature — Zombie Wizard {B}{B}, 2/2 (EDHREC rank
// 1842):
//
//	"Whenever this creature or another Zombie you control dies, you
//	 draw a card and you lose 1 life."
//
// The Zombie deck's card draw. One dies trigger with two conditions:
// the Augur's own death (the card is already in the graveyard when
// the harvester finds it, and cardDied reads it by ID) or a Zombie
// its controller controlled dying. Whether it was a Zombie is read as
// it last existed on the battlefield (CR 603.10a, #1679), so a
// changeling counts and so does a creature that was a Zombie only
// through a grant — Maskwood Nexus, a lord's type grant. Draw first,
// lose life second, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c26887e1-27f4-4550-921b-53460e43c079",
		Name:         "Undead Augur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B17SelfOrZombieYouControlDied(ev, source, g)
			}, "Undead Augur — draw a card and lose 1 life", effects.Do(
				// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -1)
				}),
			)),
		},
	})
}
