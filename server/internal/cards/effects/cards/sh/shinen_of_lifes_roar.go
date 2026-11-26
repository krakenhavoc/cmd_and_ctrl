package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shinen of Life's Roar — {1}{G} Creature — Spirit 1/2 (Archenemy):
//
//	"All creatures able to block this creature do so.
//	 Channel — {2}{G}{G}, Discard this card: All creatures able to block
//	 target creature this turn do so."
//
// The static is Prized Unicorn's lure; the channel is a hand-zone activated
// ability (Boseiju) whose effect is Taunting Challenge's turn-long lure.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b9b3347-a374-40fd-a12a-73187225194e",
		Name:         "Shinen of Life's Roar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {2}{G}{G}, Discard this card: All creatures able to block target creature this turn do so.",
			Cost:    game.AbilityCost{Mana: "{2}{G}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.BlockRequirementUntilEOT{
						Target: t.ID,
						Kind:   game.BlockRequirementLure,
						Label:  "Shinen of Life's Roar — all creatures able to block it do so",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
