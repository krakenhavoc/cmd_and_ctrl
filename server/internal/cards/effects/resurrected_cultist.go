package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resurrected Cultist — Creature — Human Cleric {2}{B}, 4/1 (Duskmourn:
// House of Horror, #4466):
//
//	"Delirium — {2}{B}{B}: Return this card from your graveyard to the
//	 battlefield with a finality counter on it. Activate only if there are
//	 four or more card types among cards in your graveyard and only as a
//	 sorcery. (If a creature with a finality counter on it would die,
//	 exile it instead.)"
//
// Gastal Thrillroller's graveyard return
// (returnThisFromGraveyardWithAFinalityCounter) behind Shifting Woodland's
// delirium gate. Delirium is an ability word (CR 207.2c); the "activate
// only if" is checked as the ability is activated (CR 602.5b), and the
// Cultist itself counts, as a creature card in that graveyard. The
// counter is on the creature as it enters, and it is exiled instead of
// dying (CR 122.1h). Sorcery timing is CR 602.5d.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b6d37a7-a88e-45dd-b45e-642a44caaf05",
		Name:         "Resurrected Cultist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Delirium — {2}{B}{B}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if there are four or more card types among cards in your graveyard and only as a sorcery.",
			Cost:         ManaCost("{2}{B}{B}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return b16CardTypesInGraveyard(g, controller) >= 4
			},
			Effect: returnThisFromGraveyardWithAFinalityCounter,
		}},
	})
}
