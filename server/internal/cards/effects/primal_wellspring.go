package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Wellspring — Land, the back face of Primal Amulet
// (primal_amulet.go), registered under "<oracle_id>#1":
//
//	"({T}: Add one mana of any color. When that mana is spent to cast an
//	 instant or sorcery spell, copy that spell and you may choose new
//	 targets for the copy.)"
//
// The parenthetical is reminder text for the ability itself, so it is
// the whole card. The mana is the same any-colour pick Birds of
// Paradise makes, and the copy is a spend rider (#1547) with Pyromancer's
// Goggles' body: the trigger goes on the stack above the spell the mana
// paid for, so the copy resolves first; the spell is named on the
// item's payload and copied from last-known information if it was
// countered in response; the copy is not cast, so it spent no mana and
// does not trigger the Wellspring again. The filter is a cast of an
// instant OR a sorcery, in any colour.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:       primalAmuletOracleID + "#1",
		Name:           "Primal Wellspring",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Primal Wellspring", game.ManaSpendTrigger{
				Label:  "Primal Wellspring — copy that spell",
				Effect: copyTheSpellYouJustCast,
			}, ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery"))},
		}},
	})
}
