package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mark of Sakiko — Enchantment — Aura {1}{G} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature deals combat damage to a player, add
//	 that much {G}. Until end of turn, you don't lose this mana as steps and phases
//	 end.""
//
// A granted triggered ability (ADR 0093) on the enchanted creature: it belongs to the
// creature, so "you" is the creature's controller, who gets the mana. The amount is the
// damage that event carried (first strike and regular damage are two events); the mana
// carries the keep-until-end-of-turn rider (KeepManaUntilEndOfTurn), as Brazen
// Collector's does.
//
// No simplification.
const markOfSakikoGrant = "mark-of-sakiko/add-green"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "18a98d44-54ee-4d6d-ab32-27e294ac8995",
		Name:         "Mark of Sakiko",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: markOfSakikoGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Mark of Sakiko — add that much {G}, kept until end of turn", effects.MarkOfSakikoMana),
			},
			Text: "Whenever this creature deals combat damage to a player, add that much {G}. Until end of turn, you don't lose this mana as steps and phases end.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(markOfSakikoGrant)},
	})
}
