package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Miser — Creature — Zombie Wizard {4}{B}, 4/4:
//
//	"Whenever you discard a creature card, create a 2/2 black Zombie
//	 creature token.
//	 Whenever you discard a land card, add {B}{B}.
//	 Whenever you discard a noncreature, nonland card, draw a card."
//
// Three separate triggered abilities on EventDiscardCard, one per
// printed sentence, each fired once per card discarded. The card is
// read where it now sits (graveyard, or the top of a library under
// Library of Leng, or exile under Rest in Peace): the discard happened
// either way (CR 701.9a), so a destination replacement leaves every
// payoff intact.
//
// A card that is both a creature and a land (Dryad Arbor) sets off the
// first two abilities and not the third; a card that is neither a
// creature nor a land sets off only the third.
//
// "Add {B}{B}" is an ordinary triggered ability that goes on the stack
// (it is not a triggered mana ability, CR 605.1b: nothing here
// resolves a mana ability), so the mana lands in the pool when the
// trigger resolves and empties at the end of that step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3984e5f6-8643-4104-aa9c-18d474e07801",
		Name:         "Bone Miser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventDiscardCard, youDiscardedA(game.Card.IsCreature),
				"Bone Miser — create a 2/2 black Zombie",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
				game.Purpose{DiscardPayoff: &game.DiscardPayoff{Types: []string{"creature"}, Tokens: 1}}),
			On(game.EventDiscardCard, youDiscardedA(game.Card.IsLand),
				"Bone Miser — add {B}{B}",
				Do(AddMana{Produced: "{B}{B}"})),
			On(game.EventDiscardCard, youDiscardedA(func(c game.Card) bool { return !c.IsCreature() && !c.IsLand() }),
				"Bone Miser — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

// youDiscardedA is the condition of Bone Miser's three abilities:
// source's controller discarded a card that match accepts.
func youDiscardedA(match func(game.Card) bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b30YouDiscardedCardWhere(ev, source, g, match)
	}
}
