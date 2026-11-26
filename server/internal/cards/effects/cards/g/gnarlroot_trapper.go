package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarlroot Trapper — Creature — Elf Druid {B}, 1/1 (Magic Origins, S82 older
// sets):
//
//	"{T}, Pay 1 life: Add {G}. Spend this mana only to cast an Elf creature spell.
//	 {T}: Target attacking Elf you control gains deathtouch until end of turn."
//
// The first ability is Blightsoil Druid's mana ability (a tap and 1 life, CR
// 605.1a) with Bucolic Ranch's spend restriction, narrowed to creature spells
// that are Elves: the mana pays for nothing else (CR 106.6). The second is an
// ordinary tap ability on an attacking Elf you control, the Trapper itself
// included when it attacks; the type is read through the layered view. Both
// taps are subject to summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3cc22ea-2ccf-409b-b471-51adb63c7fda",
		Name:         "Gnarlroot Trapper",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced:     "{G}",
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSubtype("Elf")},
			Label:        "{T}, Pay 1 life: Add {G}. Spend this mana only to cast an Elf creature spell.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target attacking Elf you control gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking Elf you control", effects.AttackingCreature(), effects.YouControl(), effects.OfSubtype("Elf")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}, Label: "Gnarlroot Trapper — deathtouch"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
