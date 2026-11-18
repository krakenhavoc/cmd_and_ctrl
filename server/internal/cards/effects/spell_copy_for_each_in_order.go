package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// copySpellForEachChosenInOrder is the keyed answer shared by Precursor
// Golem, Mirrorwing Dragon and Zada, Hedron Grinder: "that player copies
// that spell for each other <creature> … Each copy targets a different
// one of those", with the copies made in the order the player put the
// creatures in. The carry is the spell and the object the spell targets;
// a creature that is no longer a legal target by the answer gets no copy.
// The copies are ONE copy instruction (#5918, CR 707.10e), so a
// replacement of it (Twinning Staff) applies once.
//
// Unanswered, the player has left the game, and so would the copies they
// controlled (CR 800.4a).
func copySpellForEachChosenInOrder(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 || !r.Answered {
		return nil
	}
	order := r.Picked
	if len(order) != len(r.Candidates) {
		order = r.Candidates
	}
	g.CopySpellRetargetedEachForEffect(r.Carry[0], r.Chooser, r.Carry[1], order)
	return nil
}
