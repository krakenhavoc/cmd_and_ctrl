package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f034137e-c19f-48af-a2f2-f767bf87a275",
		Name:         "Draconic Destiny",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: draconicDestinyGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Draconic Destiny — +1/+0 until end of turn"),
			}},
			Text: "{1}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying", "haste"),
			effects.GrantAbilitiesToAttached(draconicDestinyGrant),
			effects.AttachedIsAlso("Dragon"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Draconic Destiny — return this card to its owner's hand", effects.EndlessEvilReturn),
		},
	})
}
