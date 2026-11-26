package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Grafstone — {2} Artifact (Shadows over Innistrad, S82 OLD b293):
//
//	"This artifact enters tapped.
//	 {T}: Choose a color of a card in your graveyard. Add one mana of that
//	 color."
//
// Mox Amber's derived mana over the colours of the cards in your graveyard:
// the pipe string offers each colour once and the player picks. A graveyard
// with no coloured card offers nothing, so the artifact taps and adds no mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c702b96-3288-435d-9154-5b7454419896",
		Name:         "Corrupted Grafstone",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.OldB293ProducedFromGraveyardColors(),
			Label:        "Choose a color of a card in your graveyard. Add one mana of that color",
		}},
	})
}
