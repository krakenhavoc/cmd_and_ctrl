package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Biophagus — Creature — Human Tyranid Wizard {1}{G}, 1/3 (EDHREC rank
// 2980):
//
//	"Genomic Enhancement — {T}: Add one mana of any color. If this mana
//	 is spent to cast a creature spell, that creature enters with an
//	 additional +1/+1 counter on it."
//
// "Genomic Enhancement" is an ability word (CR 207.2c) and does
// nothing. The mana is Birds of Paradise's five-colour pick, with no
// identity narrowing (none is printed), and carries an entry spend
// rider (#1547): when the token pays for a creature spell, the rider is
// stamped on the spell's payment record, and the entry pipeline seeds
// one additional +1/+1 counter onto the creature's entry event beside
// any the card prints — so Hardened Scales and Doubling Season see it,
// and its own enters trigger finds it already there.
//
// "This mana" is one activation's mana: a doubler that makes the pick
// two tokens is still one additional counter.
//
// One declared simplification, weaker than printed: with strict mana
// off the pool is never spent (ADR 0068 §3), so no token pays and the
// creature enters without the counter.
func init() {
	Register(Spec{
		OracleID:     "7c06a368-3306-48e0-825e-a8fdc81fa343",
		Name:         "Biophagus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the creature enters without the extra +1/+1 counter."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
			SpendRiders: []game.ManaSpendRider{
				SpentEntersWithCounters("+1/+1", 1, ManaRestrictCast, ManaRestrictType("Creature")),
			},
		}},
	})
}
