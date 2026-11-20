package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "51feda69-0ae8-4df3-bb50-a4bb9deb965d",
		Name:         "Species Specialist",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Species Specialist"),
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.NamedTribe == "" {
					return false
				}
				dead, ok := diedCreature(ev, g)
				return ok && leftAsSubtype(ev, dead, source.NamedTribe)
			}, "Species Specialist — you may draw a card", Do(DrawCards{N: 1})),
				"Species Specialist — a creature of the chosen type died. Draw a card?", 1),
		},
	})
}
