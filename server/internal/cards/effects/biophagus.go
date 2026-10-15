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
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:     "7c06a368-3306-48e0-825e-a8fdc81fa343",
		Name:         "Biophagus",
		Completeness: CompletenessFull,
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
