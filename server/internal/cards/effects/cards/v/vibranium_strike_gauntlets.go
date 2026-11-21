package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vibranium Strike Gauntlets — Artifact — Equipment {4} (Marvel Super
// Heroes Commander, #4455):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +3/+0 and has trample and "Whenever this
//	 creature deals combat damage to a player, draw a card."
//	 Equip {3}"
//
// Embercleave's flash and enters-attached trigger, Conformer Shuriken's
// granted trigger bundle (ADR 0093): the draw is the equipped creature's
// own, so it is its controller who draws, and it is gone with the
// Gauntlets when they are moved or destroyed.
//
// No simplification.
const vibraniumStrikeGauntletsGrant = "vibranium-strike-gauntlets/combat-draw"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a9d36eb-db72-459a-8de9-6d04a1d7563c",
		Name:            "Vibranium Strike Gauntlets",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Grants: []effects.AbilityGrant{{
			Key: vibraniumStrikeGauntletsGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Vibranium Strike Gauntlets — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("trample"),
			effects.GrantAbilitiesToAttached(vibraniumStrikeGauntletsGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Vibranium Strike Gauntlets — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
