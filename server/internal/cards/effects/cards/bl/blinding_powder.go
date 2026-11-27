package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinding Powder — Artifact — Equipment {1} (Betrayers of Kamigawa, #5602):
//
//	"Equipped creature has "Unattach Blinding Powder: Prevent all combat damage
//	 that would be dealt to this creature this turn."
//	 Equip {2}"
//
// Razor Boomerang's shape: the equipped creature is granted an activated
// ability whose only cost unattaches the Equipment (UnattachTheGrantor, so it
// is paid as the ability is activated, with no mana and no tap), and whose
// effect is Oketra's Avenger's shield on the creature: all combat damage that
// would be dealt to it this turn, pinned to that object (CR 400.7).
//
// No simplification.
func init() {
	const grant = "blinding-powder/shield"
	effects.Register(effects.Spec{
		OracleID:     "75e13f7f-aded-405c-97ce-124fa0cb1929",
		Name:         "Blinding Powder",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "Unattach Blinding Powder: Prevent all combat damage that would be dealt to this creature this turn.",
				Purpose: effects.PreventsDamage(),
				Cost:    effects.UnattachTheGrantor(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ApplyShieldThis(g, item, effects.PreventDamageFromSource{CombatOnly: true})
				},
			}},
			Text: "Unattach Blinding Powder: Prevent all combat damage that would be dealt to this creature this turn.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
