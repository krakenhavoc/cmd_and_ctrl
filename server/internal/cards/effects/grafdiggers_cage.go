package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// Sentence two lands. It costs nothing that the zones in question are
// exactly the two ADR 0066 just opened: a flashed-back Faithless
// Looting and a Bolas's Citadel cast off the top are both refused,
// and the enumerator and the view agree because all three read the
// same gate.
//
// Sentence one does NOT land, and the caveat says so in the player's
// words. "Can't enter the battlefield" is a replacement over an
// entry, not a restriction over a cast, and it wants the CR 614 entry
// pipeline to grow a refusal — a different seam, and one nothing else
// in the catalog is waiting on yet.
func init() {
	Register(Spec{
		OracleID:     "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd",
		Name:         "Grafdigger's Cage",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only the second sentence is enforced: players can't cast spells from graveyards or libraries. Creature cards in graveyards and libraries can still enter the battlefield — reanimation and Show and Tell-style effects are not stopped.",
		},
		CastRestrictions: []game.CastRestriction{
			PlayersCantCastFrom(
				"Grafdigger's Cage — players can't cast spells from graveyards or libraries.",
				game.ZoneGraveyard, game.ZoneLibrary),
		},
	})
}
