package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mary Jane Watson — Legendary Creature — Human Performer {1}{G/W}, 2/2
// (Marvel's Spider-Man, #4460):
//
//	"Whenever a Spider you control enters, draw a card. This ability
//	 triggers only once each turn."
//
// "A Spider you control" is read on the entering object (a token counts, a
// changeling counts, Mary Jane herself would count if she were one). "Only
// once each turn" is the per-turn trigger record Curator of Sun's Creation
// reads (b11TriggeredThisTurn): a second Spider the same turn triggers
// nothing, even if the first draw was responded to. The hybrid mana cost is
// the card's printed ManaCost.
//
// No simplification.
func init() {
	const label = "Mary Jane Watson — draw a card"
	Register(Spec{
		OracleID:     "6d44c87b-8945-4c39-b034-7690cf483992",
		Name:         "Mary Jane Watson",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, maryJaneSpiderEntered(label), label,
				Do(DrawCards{N: 1})),
		},
	})
}

// maryJaneSpiderEntered is "a Spider you control enters" and "this ability
// triggers only once each turn" in one test of the entering object.
func maryJaneSpiderEntered(label string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.HasSubtype(source.TextSubtype("Spider")) && !b11TriggeredThisTurn(g, source.InstanceID, label)
	}
}
