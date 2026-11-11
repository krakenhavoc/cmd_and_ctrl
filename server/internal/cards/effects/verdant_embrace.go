package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdant Embrace — Enchantment — Aura {3}{G}{G} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has "At the beginning of each
//	 upkeep, create a 1/1 green Saproling creature token.""
//
// +3/+3 is a layer 7c pump; the quoted trigger is a granted bundle
// (ADR 0093), so "you" is the enchanted creature's controller and the
// Saprolings are theirs, on every player's upkeep. It ends when the Aura
// leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0bbd988-50b7-476c-9183-dfabd4a95079",
		Name:         "Verdant Embrace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: verdantEmbraceSaproling,
			Triggered: []game.TriggeredAbility{
				AtEachUpkeep("Verdant Embrace — create a 1/1 green Saproling creature token",
					Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})),
			},
			Text: "At the beginning of each upkeep, create a 1/1 green Saproling creature token.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantAbilitiesToAttached(verdantEmbraceSaproling),
		},
	})
}

const verdantEmbraceSaproling = "verdant-embrace/saproling"
