package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waste Not — Enchantment {1}{B} (Shadows over Innistrad; Teenage Mutant
// Ninja Turtles Commander, #5011):
//
//	"Whenever an opponent discards a creature card, create a 2/2 black
//	 Zombie creature token.
//	 Whenever an opponent discards a land card, add {B}{B}.
//	 Whenever an opponent discards a noncreature, nonland card, draw a
//	 card."
//
// Bone Miser's three abilities turned on the opponents: one registered
// trigger per printed sentence on EventDiscardCard, each fired once per
// card discarded, the discarder being the event's Actor (Sangromancer's
// b18OpponentDiscarded). The card is read where it now sits, so a Library
// of Leng or a Rest in Peace moving the discard leaves the payoff intact
// (CR 701.9a). A creature land (Dryad Arbor) sets off the first two and
// not the third.
//
// "Add {B}{B}" is an ordinary triggered ability that goes on the stack, not
// a mana ability (CR 605.1b): the mana lands in Waste Not's controller's
// pool when it resolves and empties at the end of that step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00fdcc19-88ed-46c3-91f0-095806228105",
		Name:         "Waste Not",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, wasteNotOpponentDiscardedA(game.Card.IsCreature),
				"Waste Not — create a 2/2 black Zombie",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
			On(game.EventDiscardCard, wasteNotOpponentDiscardedA(game.Card.IsLand),
				"Waste Not — add {B}{B}",
				Do(AddMana{Produced: "{B}{B}"})),
			On(game.EventDiscardCard, wasteNotOpponentDiscardedA(func(c game.Card) bool { return !c.IsCreature() && !c.IsLand() }),
				"Waste Not — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

// wasteNotOpponentDiscardedA is the condition of Waste Not's three
// abilities: an opponent of the source's controller discarded a card that
// match accepts.
func wasteNotOpponentDiscardedA(match func(game.Card) bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !b18OpponentDiscarded(ev, source) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && match(c)
	}
}
