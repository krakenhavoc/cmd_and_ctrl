package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crackleburr — Creature — Elemental {1}{U/R}{U/R}, 2/2 (Eventide, S82 OLD batch
// 520, #5590):
//
//	"{U/R}{U/R}, {T}, Tap two untapped red creatures you control: This creature
//	 deals 3 damage to any target.
//	 {U/R}{U/R}, {Q}, Untap two tapped blue creatures you control: Return target
//	 creature to its owner's hand. ({Q} is the untap symbol.)"
//
// Belisarius Cawl's tap-others cost and Halo Fountain's untap-others cost, each
// beside the ability's own symbol. The two creatures are others, never the
// Crackleburr (it is tapped, or is being untapped, by its own symbol): the
// Crackleburr is both red and blue, but it can't pay its own clause. Tapping
// creatures for a cost that is not {T} ignores summoning sickness, and {Q} is
// AbilityCost.UntapSelf (CR 107.6): the Crackleburr must be tapped, and under
// its controller's control since their turn began (CR 302.6). The bounce
// targets any creature; a target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59a6cb4e-1646-42b5-b482-da641c0f7f6f",
		Name:         "Crackleburr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U/R}{U/R}, {T}, Tap two untapped red creatures you control: This creature deals 3 damage to any target.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
				Cost: game.AbilityCost{Tap: true, Mana: "{U/R}{U/R}", TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        effects.TargetPermanent("two untapped red creatures you control", effects.Creature(), effects.OfColor("R")),
					ExcludeSource: true,
					Label:         "two untapped red creatures you control",
				}},
				Targets: effects.TargetAny(),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label:   "{U/R}{U/R}, {Q}, Untap two tapped blue creatures you control: Return target creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost: game.AbilityCost{UntapSelf: true, Mana: "{U/R}{U/R}",
					TapOthers: effects.UntapTapped(2, "two tapped blue creatures you control", effects.Creature(), effects.OfColor("U"))},
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BounceChosenTarget,
			},
		},
	})
}
