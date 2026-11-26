package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noggle Hedge-Mage — Creature — Noggle Wizard {2}{U/R}, 2/2 (Eventide, S82 OLD
// batch 520, #5590):
//
//	"When this creature enters, if you control two or more Islands, you may tap
//	 two target permanents.
//	 When this creature enters, if you control two or more Mountains, you may have
//	 this creature deal 2 damage to target player or planeswalker."
//
// Duergar Hedge-Mage's shape: two printed triggers, so two registered ones. Each
// is an intervening "if" (CR 603.4) checked as the creature enters (nothing goes
// on the stack, and no target is chosen, otherwise) and again as the trigger
// resolves; the "may" is answered as it resolves (CR 603.5), after the targets
// were chosen. With both Islands and Mountains, both fire. The tap is of exactly
// two permanents, any player's; a target that left is skipped (CR 608.2b). The
// damage is dealt by the Mage, by its last-known information if it has left.
//
// No simplification.
func init() {
	hedgeMage := func(subtype, what string, spec *game.TargetSpec, effect effects.Effect) game.TriggeredAbility {
		label := "Noggle Hedge-Mage — if you control two or more " + subtype + "s, you may " + what
		t := effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype(subtype),
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype(subtype)), label, effect))
		t.Targets = spec
		return effects.Optional(t, "Noggle Hedge-Mage — "+what+"?")
	}
	effects.Register(effects.Spec{
		OracleID:     "16069935-4aeb-481f-98f2-46387775540f",
		Name:         "Noggle Hedge-Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			hedgeMage("Island", "tap two target permanents",
				effects.TargetPermanent("two target permanents").WithCount(2, 2), effects.TapEachLegalTarget),
			hedgeMage("Mountain", "have this creature deal 2 damage to target player or planeswalker",
				effects.TargetPlayerOrPlaneswalker(), effects.SourceDealsDamageToEachLegalTarget(2)),
		},
	})
}
