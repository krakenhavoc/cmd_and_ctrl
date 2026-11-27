package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost-Lit Warder — Creature — Spirit {1}{U}, 1/1 (Saviors of Kamigawa,
// S82 OLD b584, #5601):
//
//	"{3}{U}, {T}: Counter target spell unless its controller pays {2}.
//	 Channel — {3}{U}, Discard this card: Counter target spell unless its
//	 controller pays {4}."
//
// Censor's counter-unless-paid on an activated ability's spell target
// (CR 118.12): the spell's controller is asked, and a spell that left the
// stack in response is skipped (CR 608.2b). The channel row works only from
// the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf1a4b18-62fa-46f3-baeb-c0f699c9cdde",
		Name:         "Ghost-Lit Warder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{U}, {T}: Counter target spell unless its controller pays {2}.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost()),
				Targets: effects.TargetSpell("target spell"),
				Effect:  ghostLitWarderCounter("{2}"),
			},
			{
				Label:   "Channel — {3}{U}, Discard this card: Counter target spell unless its controller pays {4}.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: effects.TargetSpell("target spell"),
				Effect:  ghostLitWarderCounter("{4}"),
			},
		},
	})
}

func ghostLitWarderCounter(cost string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     cost,
					Question: "Ghost-Lit Warder — pay " + cost + " or your spell is countered",
				}.Apply(ctx)
			}
		}
		return nil
	}
}
