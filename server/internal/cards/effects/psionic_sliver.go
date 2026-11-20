package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psionic Sliver — Creature — Sliver {4}{U}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures have "{T}: This creature deals 2 damage to
//	 any target and 3 damage to itself.""
//
// A granted tap ability (ADR 0093) on every Sliver creature at the table.
// The Sliver that has it deals 2 damage to the target and then 3 to
// itself, so a Sliver with toughness 3 or less dies to its own ping as a
// state-based action (CR 704.5g). A Sliver that has left the battlefield
// by then is dealt nothing.
//
// No simplification.
const psionicSliverGrant = "psionic-sliver/ping"

func init() {
	Register(Spec{
		OracleID:     "ea79d4ee-c7b5-4687-9e33-22f98e46e8ec",
		Name:         "Psionic Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: psionicSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals 2 damage to any target and 3 damage to itself",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  tspB1PsionicPing,
			}},
			Text: "{T}: This creature deals 2 damage to any target and 3 damage to itself.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, psionicSliverGrant),
		},
	})
}
