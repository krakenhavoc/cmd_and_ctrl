package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafdigger's Cage — {1} Artifact:
//
//	"Creature cards in graveyards and libraries can't enter the
//	 battlefield.
//	 Players can't cast spells from graveyards or libraries."
//
// The ZONE-shaped cast restriction, and the one that proves the gate
// sees where a cast is coming FROM. It is here rather than Meddling
// Mage because Meddling Mage needs a choose-a-card-name prompt the
// engine does not have — Card.NamedTribe and Card.ChosenColor exist,
// a chosen NAME does not (ADR 0073 §8).
//
// Sentence two is the cast gate. It costs nothing that the zones in
// question are exactly the two ADR 0066 just opened: a flashed-back
// Faithless Looting and a Bolas's Citadel cast off the top are both
// refused, and the enumerator and the view agree because all three read
// the same gate.
//
// Sentence one is an entry restriction (#1985, entry_restriction.go): a
// creature card moving from a graveyard or a library onto the
// battlefield (reanimation, a library search, a put from the top of the
// library, a creature land played from a graveyard) is refused and the
// card stays where it was, which is what "can't enter" means (CR 101.2).
// A creature that comes from anywhere else (hand, exile, the stack) is
// not touched, and neither is a noncreature card, nor a token (it has no
// zone to come from). It was a cancelling replacement effect until
// #1985; a prohibition is not a replacement, so it no longer takes part
// in CR 616 ordering beside the entering card's own replacements.
func init() {
	Register(Spec{
		OracleID:     "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd",
		Name:         "Grafdigger's Cage",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			PlayersCantCastFrom(
				"Grafdigger's Cage — players can't cast spells from graveyards or libraries.",
				game.ZoneGraveyard, game.ZoneLibrary),
		},
		EntryRestrictions: []game.EntryRestriction{
			CardsInZonesCantEnter("Creature cards in graveyards and libraries can't enter the battlefield.",
				game.Card.IsCreature, game.ZoneGraveyard, game.ZoneLibrary),
		},
	})
}
