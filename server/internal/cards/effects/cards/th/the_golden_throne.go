package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Golden Throne — Legendary Artifact {4} (Warhammer 40,000 Commander):
//
//	"Arcane Life-support — If you would lose the game, instead exile The
//	 Golden Throne and your life total becomes 1.
//	 A Thousand Souls Die Every Day — {T}, Sacrifice a creature: Add three
//	 mana in any combination of colors."
//
// Arcane Life-support is a replacement of losing the game (CR 614.1a,
// #6096): game.ReplaceLoseGameExileThisLifeBecomes, read off the
// battlefield at every loss but a concession. One application replaces
// every loss one state-based check would deal (CR 704.7); the life total
// becomes 1 by gaining or losing the difference (CR 119.5). A player still
// at ten poison counters loses at the next check, with no Throne left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a68f41a2-37ee-46e1-bf3c-164c78b003ca",
		Name:         "The Golden Throne",
		Completeness: effects.CompletenessFull,
		GameActionReplacements: []game.GameActionReplacement{
			{Kind: game.ReplaceLoseGameExileThisLifeBecomes, Life: 1},
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeACreature().SacrificeOther,
			},
			Produced: effects.AnyCombinationOfColors(3),
			Label:    "A Thousand Souls Die Every Day — {T}, Sacrifice a creature: Add three mana in any combination of colors.",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
