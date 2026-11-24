package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bewitching Leechcraft — Enchantment — Aura {1}{U}:
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature has "If this creature would untap during your untap
//	 step, remove a +1/+1 counter from it instead. If you do, untap it."
//	 (Otherwise, it doesn't untap.)"
//
// The entry trigger is Claustrophobia's: it taps the creature the Aura
// enchants as it resolves, or the one it last enchanted if the Aura has
// left (CR 608.2h).
//
// The granted ability is a layer-6 grant of a bundle carrying the #6095
// untap replacement (game.UntapCounterRemoval): the enchanted creature has
// it while it has the grant, so a later "loses all abilities" takes it
// (CR 613.6), and it replaces only an untap during the untap step of the
// creature's controller (CR 502.3). With a +1/+1 counter the counter is
// removed and the creature untaps; with none it stays tapped. Two of them
// each replace the untap the other leaves, so the creature then needs two
// counters to untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5d803bc-7fd4-4570-8cd6-cb138727a66e",
		Name:         "Bewitching Leechcraft",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key:  bewitchingLeechcraftGrant,
			Text: "If this creature would untap during your untap step, remove a +1/+1 counter from it instead. If you do, untap it.",
			UntapCounterRemovals: []game.UntapCounterRemoval{
				effects.RemoveAPlusOneCounterInsteadOfUntapping("Bewitching Leechcraft — remove a +1/+1 counter instead of untapping"),
			},
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(bewitchingLeechcraftGrant)},
		Triggered: []game.TriggeredAbility{effects.WhenThisEnters("Bewitching Leechcraft — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)},
	})
}

const bewitchingLeechcraftGrant = "bewitching-leechcraft/remove-a-counter-to-untap"
