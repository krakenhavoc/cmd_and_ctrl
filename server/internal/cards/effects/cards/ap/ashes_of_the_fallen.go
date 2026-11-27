package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashes of the Fallen — Artifact {2} (Saviors of Kamigawa, S82 OLD b593):
//
//	"As this artifact enters, choose a creature type.
//	 Each creature card in your graveyard has the chosen creature type in addition to
//	 its other types."
//
// Rukarumel's card half (Leyline of Transformation's off-battlefield layer-4 static,
// ADR 0039 amendment 2026-10-15) narrowed to the graveyard: the type is appended, the
// others kept, and nothing is added before a type is chosen. The type is the stored
// "as this enters" choice (CR 614.12).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6437867b-a658-4a71-a8a5-0c5a782d1f84",
		Name:         "Ashes of the Fallen",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Ashes of the Fallen — choose a creature type"),
		Static: []game.StaticAbility{{
			Label:        "Each creature card in your graveyard has the chosen creature type in addition to its other types.",
			Layer:        game.Layer4Type,
			AffectsZones: []game.ZoneKind{game.ZoneGraveyard},
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source.NamedTribe != "" && target.Owner == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				effects.AddChosenTribe(c, source)
			},
		}},
	})
}
