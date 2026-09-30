package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umbral Collar Zealot — Creature — Human Cleric {1}{B}, 3/2 (EDHREC
// rank 4488):
//
//	"Sacrifice another creature or artifact: Surveil 1."
//
// An ordinary CR 602 sacrifice-outlet ability with no mana or tap
// component, so it can fire any number of times a turn, at instant
// speed, for as long as there is something else to feed it.
//
// Declared simplification: "another" is enforced by NAME
// (b03NotNamed), the catalog's standing convention for an activated
// ability's target/cost clause — ActivatedAbility has no per-source
// rewrite the way a triggered ability's TargetsFrom does (see Kelpie
// Guide, Thassa Deep-Dwelling). The Zealot cannot sacrifice itself
// (correct — it is a creature, and the printed "another" excludes
// it), but it also cannot sacrifice a SECOND Umbral Collar Zealot,
// which the printed card allows.
func init() {
	Register(Spec{
		OracleID:     "12db6263-75c2-442f-a1a5-7af7915f8f9f",
		Name:         "Umbral Collar Zealot",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The sacrifice cost can't be paid with another Umbral Collar Zealot."},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice another creature or artifact: Surveil 1.",
			Cost: game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()), b03NotNamed("Umbral Collar Zealot")),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
