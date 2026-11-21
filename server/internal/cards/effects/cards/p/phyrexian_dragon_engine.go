package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Dragon Engine — Artifact Creature — Phyrexian Dragon {3}, 2/2
// (The Brothers' War, #5482):
//
//	"Double strike
//	 When this creature enters from your graveyard, you may discard your
//	 hand. If you do, draw three cards.
//	 Unearth {3}{R}{R}
//	 (Melds with Mishra, Claimed by Gix.)"
//
// The enters trigger reads where the creature came from and whose
// graveyard it was (EventETB.EnteredFrom and EnteredFromOwner), so it
// fires for an unearth or any other return from its controller's own
// graveyard and never when it is cast or comes from another player's
// graveyard. The "you may" is a keyed yes/no; a yes discards the whole
// hand, even an empty one, and draws three. The reminder line is Mishra,
// Claimed by Gix's meld ability and needs nothing here.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "022b42a6-ee0f-4080-b973-c1b68938c3be",
		Name:            "Phyrexian Dragon Engine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.EnteredFromYourGraveyard),
				"Phyrexian Dragon Engine — you may discard your hand; if you do, draw three cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Phyrexian Dragon Engine — discard your hand? If you do, draw three cards.",
						YesLabel: "Discard and draw three",
						NoLabel:  "Keep your hand",
						Then:     effects.S82BroB20DragonEngineDiscard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{R}{R}")},
	})
}
