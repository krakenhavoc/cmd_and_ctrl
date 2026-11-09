package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chronicle of Victory — Legendary Artifact {6} (Lorwyn Eclipsed, #4458):
//
//	"As Chronicle of Victory enters, choose a creature type.
//	 Creatures you control of the chosen type get +2/+2 and have first
//	 strike and trample.
//	 Whenever you cast a spell of the chosen type, draw a card."
//
// Vanquisher's Banner's shape: the CR 614.12 as-enters choice stamped
// on the permanent (Card.NamedTribe), a tribal anthem and two tribal
// keyword grants read through that choice. The cast trigger is Door of
// Destinies' "a spell of the chosen type", not "a creature spell", so a
// kindred spell or a changeling of the type counts; it resolves before
// the spell (the ruling), and still draws if the spell is countered.
//
// No simplification.
func init() {
	yours := TribeFilter{Chosen: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "e949b362-4988-4482-ac51-c1049af9040c",
		Name:         "Chronicle of Victory",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Chronicle of Victory"),
		Static: []game.StaticAbility{
			TribalAnthem(yours, 2, 2),
			TribalKeywordGrant(yours, "first strike"),
			TribalKeywordGrant(yours, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || source.NamedTribe == "" {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasSubtype(source.NamedTribe)
			}, "Chronicle of Victory — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
