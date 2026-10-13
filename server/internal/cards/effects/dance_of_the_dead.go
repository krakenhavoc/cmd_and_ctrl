package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dance of the Dead — Enchantment — Aura {1}{B}:
//
//	"Enchant creature card in a graveyard
//	 When this Aura enters, if it's on the battlefield, it loses
//	 "enchant creature card in a graveyard" and gains "enchant creature
//	 put onto the battlefield with this Aura." Put enchanted creature
//	 card onto the battlefield tapped under your control and attach this
//	 Aura to it. When this Aura leaves the battlefield, that creature's
//	 controller sacrifices it.
//	 Enchanted creature gets +1/+1 and doesn't untap during its
//	 controller's untap step.
//	 At the beginning of the upkeep of enchanted creature's controller,
//	 that player may pay {1}{B}. If the player does, untap that
//	 creature."
//
// ADR 0154 (#1780): Animate Dead's seam with the return tapped
// (AnimateEnchantedCard{Tapped: true}). The upkeep trigger watches the
// upkeep of whoever controls the enchanted creature now, and that player
// is asked to pay, through the ordinary "you may pay" prompt. With two
// on one creature each gives +1/+1 and each upkeep cost may be paid
// (2004-10-04 ruling), which the per-Aura statics and triggers give.
//
// Like every catalogued Aura but Animate Dead it is marked Full: no Aura
// can yet choose what it enchants when it is put onto the battlefield
// without being cast (CR 303.4f, #3158), and the cards that do that
// (Replenish, Brilliant Restoration) carry that caveat themselves. ADR
// 0154 Question 1 gave Animate Dead its own caveat because players reach
// for Sun Titan and Worldgorger Dragon lines with it first.
func init() {
	Register(Spec{
		OracleID:              "15d5b198-edbe-40f3-87e2-d72f7f0c58ab",
		Name:                  "Dance of the Dead",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreatureCardInGraveyard(),
		Static:                []game.StaticAbility{PumpAttached(1, 1)},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dance of the Dead — put enchanted creature card onto the battlefield tapped under your control",
				Do(AnimateEnchantedCard{Tapped: true})),
			On(game.EventBeginUpkeep, UpkeepOfEnchantedCreaturesController,
				"Dance of the Dead — that player may pay {1}{B} to untap that creature",
				mayPayToUntapEnchantedCreature("Dance of the Dead", "{1}{B}")),
		},
	})
}
