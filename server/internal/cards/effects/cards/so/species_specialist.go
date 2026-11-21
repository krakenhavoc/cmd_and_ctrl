package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Species Specialist — Creature — Human Warrior {2}{B}{B}, 2/3 (Commander
// 2020, #5513):
//
//	"As this creature enters, choose a creature type.
//	 Whenever a creature of the chosen type dies, you may draw a card."
//
// The choice is Kindred Discovery's ChooseCreatureTypeAsEnters, stored on the
// permanent (Card.NamedTribe). The dies trigger watches every creature,
// anyone's, and reads its type as it last existed on the battlefield
// (leftAsSubtype, CR 603.10a), so a changeling or a creature a lord made that
// type counts. The draw is a "you may" asked as the trigger resolves (CR
// 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51feda69-0ae8-4df3-bb50-a4bb9deb965d",
		Name:         "Species Specialist",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Species Specialist"),
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.NamedTribe == "" {
					return false
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftAsSubtype(ev, dead, source.NamedTribe)
			}, "Species Specialist — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Species Specialist — a creature of the chosen type died. Draw a card?", 1),
		},
	})
}
