package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const draconicDestinyGrant = "draconic-destiny/firebreathing" // an on-disk identity: never renamed

// Draconic Destiny — Enchantment — Aura {1}{R}{R} (The Brothers' War):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying, haste, and "{1}: This
//	 creature gets +1/+0 until end of turn." It's a Dragon in addition to
//	 its other types.
//	 When enchanted creature dies, return this card to its owner's hand."
//
// Angelic Destiny's shape (additive type, return trigger) with Fallen
// Ideal's granted activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f034137e-c19f-48af-a2f2-f767bf87a275",
		Name:         "Draconic Destiny",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: draconicDestinyGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Draconic Destiny — +1/+0 until end of turn"),
			}},
			Text: "{1}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying", "haste"),
			GrantAbilitiesToAttached(draconicDestinyGrant),
			AttachedIsAlso("Dragon"),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Draconic Destiny — return this card to its owner's hand", endlessEvilReturn),
		},
	})
}
