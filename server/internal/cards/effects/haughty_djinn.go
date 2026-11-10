package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haughty Djinn — Creature — Djinn {1}{U}{U}, */4 (Foundations Commander;
// first printed in Kaladesh):
//
//	"Flying
//	 Haughty Djinn's power is equal to the number of instant and sorcery
//	 cards in your graveyard.
//	 Instant and sorcery spells you cast cost {1} less to cast."
//
// The power is a characteristic-defining ability (layer 7a) counting the
// instant and sorcery cards in its CONTROLLER's graveyard; a split card
// counts once. The discount is Archmage of Runes's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44036bb5-f300-40bd-9180-6b246315d886",
		Name:            "Haughty Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Haughty Djinn's power is equal to the number of instant and sorcery cards in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = cmdrB17InstantSorceryCardsInGraveyard(g, source.Controller)
			},
		}},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
	})
}
