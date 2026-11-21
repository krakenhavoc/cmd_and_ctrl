package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shielded Aether Thief — Creature — Vedalken Rogue {1}{U}, 0/4:
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Whenever this creature blocks, you get {E} (an energy counter).
//	 {T}, Pay {E}{E}{E}: Draw a card."
//
// ADR 0129 PR 1 (#1995). "Whenever this creature blocks" has no
// object, so it triggers once however many attackers it blocks
// (CR 509.3a, selfBlocksOnce).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "514df71a-138f-460f-a03c-bd1af8e146e2",
		Name:            "Shielded Aether Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Shielded Aether Thief — you get {E}", effects.EbYouGetEnergy(1)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}{E}{E}: Draw a card.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(3)),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect:  effects.EbDrawACard,
		}},
	})
}
