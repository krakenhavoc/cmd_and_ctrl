package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokenzan Spellblade — Creature — Ogre Samurai Shaman {4}{R}, 2/3 (Saviors
// of Kamigawa, #5601):
//
//	"Bushido 1 (Whenever this creature blocks or becomes blocked, it gets
//	 +1/+1 until end of turn.)
//	 {1}{R}: This creature gets +X/+0 until end of turn, where X is the
//	 number of cards in your hand."
//
// Bushido is the engine's (CR 702.45a). X is counted as the ability resolves
// (CR 608.2h); the ability can be activated any number of times, each
// counting the hand again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26e47cc5-9dc7-4d54-80a6-1eb77587df36",
		Name:         "Sokenzan Spellblade",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(1, "Sokenzan Spellblade")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +X/+0 until end of turn, where X is the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.SokB1CardsInYourHand(ctx)
				if x <= 0 {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Sokenzan Spellblade — +X/+0"}.Apply(ctx)
			},
		}},
	})
}
