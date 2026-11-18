package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clawing Torment — Enchantment — Aura {B} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets -1/-1 and can't
//	 block.
//	 Enchanted permanent has "At the beginning of your upkeep, you lose 1
//	 life.""
//
// The pump and the block restriction wait for the host to be a creature; the
// quoted upkeep trigger is a granted bundle (ADR 0093) on the permanent
// whatever it is, so "you" is the enchanted permanent's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ac5c149-f6af-46b6-974e-14689e9a583d",
		Name:         "Clawing Torment",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []AbilityGrant{{
			Key: clawingTormentUpkeep,
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Clawing Torment — you lose 1 life", Do(GainLife{Amount: -1})),
			},
			Text: "At the beginning of your upkeep, you lose 1 life.",
		}},
		Static: []game.StaticAbility{
			b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, -1, -1),
			neoB12CantBlockWhenCreature(),
			GrantAbilitiesToAttached(clawingTormentUpkeep),
		},
	})
}

const clawingTormentUpkeep = "clawing-torment/upkeep"
