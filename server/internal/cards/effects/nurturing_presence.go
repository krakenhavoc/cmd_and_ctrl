package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nurturing Presence — Enchantment — Aura {1}{W} (Innistrad: Crimson Vow):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever a creature you control enters, this
//	 creature gets +1/+1 until end of turn."
//	 When this Aura enters, create a 1/1 white Spirit creature token with
//	 flying."
//
// The granted trigger is Cool Fluffy Loxodon's entering condition (any
// creature under the enchanted creature's controller's control, itself
// included) with a +1/+1 until end of turn on the enchanted creature as the
// grant's source. The Aura's own enters trigger makes the Spirit as it
// arrives, whatever it enchants.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f03c834-2cb3-4c54-bee6-c64f3a5fce47",
		Name:         "Nurturing Presence",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: nurturingPresencePump,
			Triggered: []game.TriggeredAbility{
				On(game.EventETB, CreatureEnteredUnderYourControl,
					"Nurturing Presence — this creature gets +1/+1 until end of turn",
					thisGetsUntilEndOfTurn(1, 1, "Nurturing Presence — +1/+1 until end of turn")),
			},
			Text: "Whenever a creature you control enters, this creature gets +1/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(nurturingPresencePump)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nurturing Presence — create a 1/1 white Spirit creature token with flying",
				Do(vowB12SpiritToken())),
		},
	})
}

const nurturingPresencePump = "nurturing-presence/pump"
