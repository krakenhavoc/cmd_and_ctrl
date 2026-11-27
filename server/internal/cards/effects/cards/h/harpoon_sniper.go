package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harpoon Sniper — Creature — Merfolk Archer {2}{W}, 2/2 (Lorwyn):
//
//	"{W}, {T}: This creature deals X damage to target attacking or blocking
//	 creature, where X is the number of Merfolk you control."
//
// Arcbound Javelineer's target (an attacking or blocking creature) with an X
// counted as the ability resolves (CR 608.2h): the Merfolk its controller has
// then, the Sniper included if it is still there (it is a Merfolk), read
// through the Sniper's text (CR 612.2). A target that has left combat is
// illegal (CR 608.2b). The Sniper deals the damage as it last existed if it
// has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2856844d-f727-48aa-8387-cb6079afb40c",
		Name:         "Harpoon Sniper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: This creature deals X damage to target attacking or blocking creature, where X is the number of Merfolk you control.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				merfolk := ctx.TextSubtype("Merfolk")
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.HasSubtype(merfolk) {
						n++
					}
				}
				return n
			}),
		}},
	})
}
