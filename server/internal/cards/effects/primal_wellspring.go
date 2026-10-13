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
// One declared simplification, weaker than printed and shared with every
// spend rider: with an unrecorded payment the pool is never spent (ADR
// 0068 §3), so no token pays and the spell is never copied.
func init() {
	Register(Spec{
		OracleID:     primalAmuletOracleID + "#1",
		Name:         "Primal Wellspring",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When mana payment is skipped, the game doesn't see which mana you spent, so the spell is never copied.",
		},
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
