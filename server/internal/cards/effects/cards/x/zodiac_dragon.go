package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zodiac Dragon — Creature — Dragon {7}{R}{R}, 8/8
// (Portal Three Kingdoms, #5622):
//
//	"When this creature is put into your graveyard from the battlefield, you may return it to your hand."
//
// "Put into your graveyard" is a trip from the battlefield to the graveyard of
// the Dragon's controller: its owner, when it is also its controller (a
// stolen one goes to its owner's graveyard, which is not "your" graveyard).
// The return only moves the card while it is still in that graveyard as the
// trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5c2cf0f-c1c0-4fd6-9b5f-7d4aa25246a7",
		Name:         "Zodiac Dragon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source) && source.Owner == source.Controller
			}, "Zodiac Dragon — return it to your hand", effects.ReturnThisCardFromYourGraveyardToYourHand),
				"Zodiac Dragon — return it to your hand?"),
		},
	})
}
