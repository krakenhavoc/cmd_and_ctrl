package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basal Sliver — Creature — Sliver {2}{B}, 2/2 (Time Spiral, #5596):
//
//	"All Slivers have "Sacrifice this permanent: Add {B}{B}.""
//
// A granted MANA ability (ADR 0093, Gemhide Sliver's shape) on every
// Sliver at the table: sacrificing the Sliver that has it adds {B}{B} to
// its controller's pool. It is a mana ability, so it doesn't use the
// stack.
//
// No simplification.
const basalSliverGrant = "basal-sliver/mana"

func init() {
	Register(Spec{
		OracleID:     "18e7ba0d-88a5-466b-9228-d3f5aae8f83e",
		Name:         "Basal Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: basalSliverGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Sacrifice: true},
				Produced: "{B}{B}",
				Label:    "Sacrifice this permanent: Add {B}{B}",
			}},
			Text: "Sacrifice this permanent: Add {B}{B}.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, basalSliverGrant),
		},
	})
}
