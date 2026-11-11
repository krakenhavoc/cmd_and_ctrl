package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feather, Radiant Arbiter — Legendary Creature — Angel {R}{W}{W}, 4/3
// (Murders at Karlov Manor Commander):
//
//	"Flying, lifelink
//	 Whenever you cast a noncreature spell that targets only Feather, you
//	 may choose any number of other creatures that spell could target and
//	 pay {2} for each of those creatures. If you do, for each of those
//	 creatures, copy that spell. The copy targets that creature. (Copies of
//	 permanent spells become tokens.)"
//
// "Targets only Feather" is every target of the spell being Feather
// (StackItemSoleTargetForEffect, Mirrorwing Dragon's read). The creatures are
// any other creatures the spell could target, whoever controls them
// (game.SpellCouldTargetInsteadForEffect rules out hexproof, protection and a
// clause that excludes one). The trigger's controller pays {2} for each (the
// pay-X prompt, Hawkeye's seam, with {2} as the repeated symbol, capped at
// the creatures there are) and then picks which, so a payment short of the
// creatures asks which to copy it for. Each copy (CR 707.10) is the caster's,
// is not cast, and has that creature in place of Feather; an Aura copy
// enchants it and a permanent copy becomes a token. A spell countered before
// the trigger resolves is still copied from last-known information (CR
// 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "957413ae-e756-4afe-886a-b57bf59d5f8d",
		Name:            "Feather, Radiant Arbiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, cmdr2S41FeatherTargetsOnlyThis,
				"Feather, Radiant Arbiter — you may pay {2} for each other creature that spell could target; copy that spell for each",
				cmdr2S41FeatherOffer),
		},
	})
}
