package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ballista Squad — Creature — Human Rebel {3}{W}, 2/2 (Mercadian Masques,
// S82 OLD, #5619):
//
//	"{X}{W}, {T}: This creature deals X damage to target attacking or blocking
//	 creature."
//
// Crimson Hellkite's X damage with Righteous Blow's target: the creature must be
// attacking or blocking as the ability is activated and again as it resolves
// (CR 608.2b). The {T} is the Squad's own, so it is summoning-sick the turn it
// arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5973725c-6e1f-4d7e-bcef-3d867ea4c244",
		Name:         "Ballista Squad",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{W}, {T}: This creature deals X damage to target attacking or blocking creature.",
			Cost:    effects.Plus(effects.ManaCost("{X}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CmdrB19DealXToTheTarget(effects.NewContext(g, item))
			},
		}},
	})
}
