package game

import "github.com/google/uuid"

// villainous_choice.go — the "faces a villainous choice again" static
// (CR 701.55c, #3140).
//
// "[A player] faces a villainous choice — [option A], or [option B]"
// is the keyword action (CR 701.55a): the player chooses one of the two
// options and every action in it is performed. The question itself is
// an ordinary option pick, queued by effects.FaceVillainousChoice; the
// only thing the engine owns is how many times each player is asked.
//
// CR 701.55c: "A replacement effect may replace an instruction to face
// a villainous choice with an instruction to face that choice some
// number of additional times. In that case, the entire process
// described in rule 701.55a is performed for that player the
// appropriate number of times one at a time." The Valeyard is the one
// card that does it: "If an opponent would face a villainous choice,
// they face that choice an additional time."
//
// It is modelled as a count read when the choice is faced, the way an
// extra vote is (ExtraVote, council_vote.go): each such permanent on
// the battlefield under an opponent's control adds one. Two Valeyards
// give three questions, which is what two replacement effects give —
// the second applies to the modified instruction, because CR 614.5
// only stops one effect applying to the same event twice.

// CatalogExtraVillainousChoice reports whether a battlefield permanent
// with this catalog key makes its controller's opponents face a
// villainous choice an additional time. Populated from
// effects.Spec.OpponentsFaceVillainousChoicesAgain. A nil hook counts
// nothing.
var CatalogExtraVillainousChoice func(key string) bool

// VillainousChoiceTimesForEffect is how many times `facing` faces one
// villainous choice right now: one, plus one for each permanent that
// makes them face it again (CR 701.55c). A permanent that has lost its
// abilities adds nothing (CR 613.1f, through CatalogAbilityKey).
// Caller must hold g.mu.
func (g *Game) VillainousChoiceTimesForEffect(facing uuid.UUID) int {
	times := 1
	if CatalogExtraVillainousChoice == nil || g.Battlefield == nil {
		return times
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == facing || c.Controller == uuid.Nil {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogExtraVillainousChoice(key) {
			times++
		}
	}
	return times
}
