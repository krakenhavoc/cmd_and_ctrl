package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const swiftReconfigurationCrewGrant = "swift-reconfiguration/crew-five"

// Swift Reconfiguration — Enchantment — Aura {W} (Neon Dynasty):
//
//	"Flash
//	 Enchant creature or Vehicle
//	 Enchanted permanent is a Vehicle artifact with crew 5 and it loses
//	 all other card types. (It's not a creature unless it's crewed.)"
//
// Sugar Coat's layer-4 shape: the card types become Artifact alone and the
// subtypes become Vehicle (the creature subtypes go with the creature type,
// CR 205.3d; supertypes stay, CR 205.4b). Unlike Sugar Coat it does not
// lose abilities, and it adds "crew 5" (Astor's grant shape, CR 702.122a)
// on top of whatever the host has. A crewed host is a creature again until
// end of turn (the crew effect is a later timestamp).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d47e820-913f-441a-a6cc-37ab3181d79a",
		Name:            "Swift Reconfiguration",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.TargetPermanent("enchant creature or Vehicle", effects.Or(effects.Creature(), effects.Subtype("Vehicle"))),
		Grants: []effects.AbilityGrant{{Key: swiftReconfigurationCrewGrant, Activated: []effects.ActivatedAbility{{
			Label:   "Crew 5",
			Cost:    effects.CrewCost(5),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Swift Reconfiguration"),
		}}, Text: "Crew 5"}},
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted permanent is a Vehicle artifact and loses all other card types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Types = []string{"Artifact"}
					c.SetSubtypes(nil)
					c.Subtypes = append(c.Subtypes, source.TextSubtypes("Vehicle")...)
				},
			},
			effects.GrantAbilities(effects.AttachedToSource, swiftReconfigurationCrewGrant),
		},
	})
}
