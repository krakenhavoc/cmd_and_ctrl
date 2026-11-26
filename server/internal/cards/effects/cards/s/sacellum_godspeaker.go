package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sacellum Godspeaker — Creature — Elf Druid {2}{G}, 2/2 (Shards of Alara, S82
// OLD batch 517, #5588):
//
//	"{T}: Reveal any number of creature cards with power 5 or greater from
//	 your hand. Add {G} for each card revealed this way."
//
// Rosheen, Roaring Prophet's mana ability that reveals any number of matching
// hand cards, none included: they are revealed as the ability resolves
// (CR 605.3b, CR 701.20a), stay in the hand, and each adds {G}. The quality is
// creature cards with printed power 5 or greater
// (game.RevealCost.PowerAtLeast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "891e9a08-35ec-422f-87a8-1f3f5c705699",
		Name:         "Sacellum Godspeaker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:    true,
				Reveal: effects.RevealAnyNumber("creature cards with power 5 or greater", game.RevealCost{CardType: "creature", PowerAtLeast: 5}),
			},
			ProducedForPaid: sacellumGodspeakerProduced,
			Label:           "{T}: Reveal any number of creature cards with power 5 or greater from your hand. Add {G} for each card revealed this way.",
		}},
	})
}

// sacellumGodspeakerProduced is "Add {G} for each card revealed this way".
func sacellumGodspeakerProduced(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"G"}, N: len(paid.Revealed)}}
}
