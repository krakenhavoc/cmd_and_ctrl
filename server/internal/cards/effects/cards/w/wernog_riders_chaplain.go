package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wernog, Rider's Chaplain — Legendary Creature — Human {W}{B}, 1/2 (Secret
// Lair Drop):
//
//	"When Wernog, Rider's Chaplain enters or leaves the battlefield, each
//	 opponent may investigate. Each opponent who doesn't loses 1 life. You
//	 investigate X times, where X is one plus the number of opponents who
//	 investigated this way.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// One trigger on both events (WhenThisEntersOrLeaves). Each opponent in turn
// order is asked once, one prompt at a time like a tempting offer (Tempt with
// Vengeance's offer chain); the opponents still to ask and then the ones who
// said yes ride the prompt's carry, split by a nil id. A yes investigates for
// that opponent, a no costs them 1 life. When the last has answered the
// controller investigates one plus the number of yeses, so with no opponents
// left in the game, once. Partner—Friends forever is a deck-construction rule
// read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cd03270-54cc-43e1-9b86-70c76960c841",
		Name:         "Wernog, Rider's Chaplain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrLeaves("Wernog, Rider's Chaplain — each opponent may investigate; you investigate one plus the number who did",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return wernogAskNext(ctx, ctx.Opponents(), nil)
				}),
		},
	})
}

// wernogAsked is one opponent's answer. Registered from init so it can name
// wernogAskNext, which names it back.
var wernogAsked game.ConfirmThen

func init() {
	wernogAsked = effects.MayAnswer("confirm/wernog-riders-chaplain", func(ctx *effects.Context, r game.ConfirmAnswer) error {
		var rest, yes []uuid.UUID
		seenSplit := false
		for _, id := range r.Carry {
			switch {
			case id == uuid.Nil:
				seenSplit = true
			case seenSplit:
				yes = append(yes, id)
			default:
				rest = append(rest, id)
			}
		}
		if r.Accepted {
			if err := (effects.Investigate{Player: r.Chooser, N: 1}).Apply(ctx); err != nil {
				return err
			}
			yes = append(yes, r.Chooser)
		} else if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -1); err != nil {
			return err
		}
		return wernogAskNext(ctx, rest, yes)
	})
}

// wernogAskNext asks the next opponent, or investigates for the controller
// once the last has answered.
func wernogAskNext(ctx *effects.Context, rest, yes []uuid.UUID) error {
	if len(rest) == 0 {
		return effects.Investigate{N: 1 + len(yes)}.Apply(ctx)
	}
	carry := append([]uuid.UUID(nil), rest[1:]...)
	carry = append(carry, uuid.Nil)
	carry = append(carry, yes...)
	return effects.MayChoice{
		Player:   rest[0],
		Question: "Wernog, Rider's Chaplain — investigate? (If you don't, you lose 1 life.)",
		Then:     wernogAsked,
		Carry:    carry,
	}.Apply(ctx)
}
