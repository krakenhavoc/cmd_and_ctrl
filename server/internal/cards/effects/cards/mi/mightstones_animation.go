package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mightstone's Animation — Enchantment — Aura {3}{U} (The Brothers' War,
// #5482):
//
//	"Enchant artifact
//	 When this Aura enters, draw a card.
//	 Enchanted artifact is a creature with base power and toughness 4/4 in
//	 addition to its other types."
//
// One continuous effect in two layers: the Creature type in layer 4 and
// base 4/4 in layer 7b (Blade of the Oni's shape), both declared
// ContinuesAfterRemoval (CR 613.6). The Aura stays attached when the
// artifact is a creature (it is still an artifact, CR 303.4d/704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de68a69a-0c4e-43b9-821d-2fe2269b5e48",
		Name:         "Mightstone's Animation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact", effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted artifact is a creature in addition to its other types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.HasFold(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
				},
			},
			effects.SetAttachedBasePT(4, 4),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mightstone's Animation — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
