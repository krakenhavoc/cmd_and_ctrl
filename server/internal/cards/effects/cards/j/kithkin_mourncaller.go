package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Mourncaller — Creature — Kithkin Scout {2}{G}, 2/2 (Lorwyn, S82
// OLD b546, #5593):
//
//	"Whenever an attacking Kithkin or Elf is put into your graveyard from
//	 the battlefield, you may draw a card."
//
// Kardur's attacking-creature death watcher: the creature was attacking, and
// was a Kithkin or an Elf, as it last existed on the battlefield (CR 603.10a;
// a changeling was both), and went to ITS OWNER's graveyard, which must be
// the Mourncaller's controller's. Its own death counts when it was attacking.
// The draw is a "may" asked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a21e773e-1f23-42af-a1bf-b92c84b7436e",
		Name:         "Kithkin Mourncaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.DiedWhileAttacking(ev, g)
					return ok && c.Owner == source.Controller &&
						(effects.LeftAsSubtype(ev, c, source.TextSubtype("Kithkin")) || effects.LeftAsSubtype(ev, c, source.TextSubtype("Elf")))
				}, "Kithkin Mourncaller — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Kithkin Mourncaller — draw a card?"),
				game.Purpose{Draws: 1}),
		},
	})
}
