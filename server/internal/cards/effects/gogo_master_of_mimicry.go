package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gogo, Master of Mimicry — Legendary Creature — Wizard {2}{U}, 2/4
// (Edea steal-and-sac deck, #1565):
//
//	"{X}{X}, {T}: Copy target activated or triggered ability you
//	 control X times. You may choose new targets for the copies. This
//	 ability can't be copied and X can't be 0. (Mana abilities can't
//	 be targeted.)"
//
// Strionic Resonator's ability-copy seam (#1223), X times. The target
// clause is AbilityOnStack narrowed to "you control" and nothing else,
// because Gogo copies activated AND triggered abilities (Lithoform
// Engine's reach, not the Resonator's). A mana ability never uses the
// stack (CR 605.3b), so the reminder text is enforced by the ability
// not being there to click.
//
// The cost is ManaCost("{X}{X}") with MinX(1): one announced X pays
// both {X}s (CR 107.3), so X = 2 costs four mana and makes two
// copies, and X = 0 is refused at announce, as printed. CopyAbility's
// Count is X, and each copy gets its own CR 707.10c re-target offer.
//
// "This ability can't be copied" is the ability's Uncopyable bit
// (#1574, ADR 0043 Decision 20). The activation carries it onto the
// stack item and game.CopyAbilityForEffect, the one door every
// ability copy comes through, refuses it: Lithoform Engine, Strionic
// Resonator, Rings of Brighthearth and another Gogo all copy nothing.
// It narrows no target clause, which is how "can't be countered"
// works too: another Gogo (a Spark Double's non-legendary copy, say)
// may target this activation, and its copies are simply not made. So
// may this Gogo, untapped, target its own earlier activation. An
// activation never targets ITSELF: its targets are chosen before the
// item is on the stack (CR 115.5).
func init() {
	Register(Spec{
		OracleID:     gogoMasterOfMimicryOracleID,
		Name:         "Gogo, Master of Mimicry",
		XMatters:     true,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:      "{X}{X}, {T}: Copy target activated or triggered ability you control X times. You may choose new targets for the copies. This ability can't be copied and X can't be 0.",
			Cost:       Plus(ManaCost("{X}{X}"), TapCost(), MinX(1)),
			Targets:    AbilityOnStack("target activated or triggered ability you control", AnAbilityYouControl()),
			Uncopyable: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok || ctx.X() < 1 {
					return nil
				}
				return CopyAbility{
					ItemID:           id,
					Controller:       item.Controller,
					Count:            ctx.X(),
					ChooseNewTargets: true,
				}.Apply(ctx)
			},
		}},
	})
}

const gogoMasterOfMimicryOracleID = "61586052-7d69-489c-84c4-0359228d131b"
