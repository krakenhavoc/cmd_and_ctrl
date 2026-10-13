package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Animate Dead — Enchantment — Aura {1}{B}:
//
//	"Enchant creature card in a graveyard
//	 When this Aura enters, if it's on the battlefield, it loses
//	 "enchant creature card in a graveyard" and gains "enchant creature
//	 put onto the battlefield with this Aura." Return enchanted creature
//	 card to the battlefield under your control and attach this Aura to
//	 it. When this Aura leaves the battlefield, that creature's
//	 controller sacrifices it.
//	 Enchanted creature gets -1/-0."
//
// ADR 0154 (#1780). The cast targets a creature card in any graveyard,
// and the Aura resolves attached to that card (CR 608.3c). Its enters
// trigger swaps its enchant clause, returns the card under its
// controller's control, attaches to the new permanent and schedules the
// delayed sacrifice (AnimateEnchantedCard, graveyard_auras.go). The
// -1/-0 is the ordinary "enchanted creature" static, which reaches
// nothing while the host is a card in a graveyard.
//
// Caveat (ADR 0154 Question 1): the engine has no CR 303.4f choice of
// what an Aura enchants when it is put onto the battlefield without being
// cast, for any Aura (#3158). Put onto the battlefield by Sun Titan,
// Replenish or a Worldgorger Dragon loop, Animate Dead enchants nothing
// and goes to the graveyard.
func init() {
	Register(Spec{
		OracleID:     "c0d8fef4-65f4-4769-982d-b397d2b7e977",
		Name:         "Animate Dead",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Put onto the battlefield any other way than by casting it (Sun Titan, Replenish), it enchants nothing and goes to the graveyard.",
		},
		Targets: EnchantCreatureCardInGraveyard(),
		Static:  []game.StaticAbility{PumpAttached(-1, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Animate Dead — return enchanted creature card to the battlefield under your control",
				Do(AnimateEnchantedCard{})),
		},
	})
}
