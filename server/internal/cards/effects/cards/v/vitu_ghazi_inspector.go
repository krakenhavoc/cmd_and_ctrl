package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vitu-Ghazi Inspector — Creature — Elf Detective {1}{G}, 1/3 (Murders at
// Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 (Exile cards with total mana value 6 or greater from your graveyard.)
//	 Reach
//	 When this creature enters, if evidence was collected, put a +1/+1
//	 counter on target creature and you gain 2 life."
//
// The offer is OptionalCollectEvidence. The enters trigger has an
// intervening if (CR 603.4) on the cast's record carried onto the
// permanent (ThisEvidenceCollected, CR 400.7d / 607.2i): put onto the
// battlefield without being cast, or cast without collecting, it does not
// trigger. The counter lands on the target and then its controller gains 2
// life; a target gone by resolution still leaves the life gain (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22a33139-a730-4792-9dd5-2c0f99f96855",
		Name:            "Vitu-Ghazi Inspector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		OptionalCosts:   []game.AdditionalCost{effects.OptionalCollectEvidence(6)},
		Triggered: []game.TriggeredAbility{effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisEvidenceCollected),
			"Vitu-Ghazi Inspector — if evidence was collected, put a +1/+1 counter on target creature and gain 2 life",
			effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }), effects.GainLife{Amount: 2})),
			effects.TargetCreature("target creature"))},
	})
}
