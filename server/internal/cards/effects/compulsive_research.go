package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compulsive Research — Sorcery {2}{U} (EDHREC rank 4386):
//
//	"Target player draws three cards. Then that player discards two
//	 cards unless they discard a land card."
//
// Three mana, three cards, and the bill comes as a land you were
// never going to play — which is why it has outlasted every other
// draw-three of its era in Commander. It is in roadmap batch 42
// (#449) under "no new machinery".
//
// "Unless they discard a land card" is a rule about the SET, not
// about the count, and no bound on the pick can express it: one card
// is a legal answer only when that card is a land, and two cards are
// a legal answer whatever they are. That is what DiscardPrompt's
// Validate hook is for — it runs on the submit path and inside the
// bot enumerator, so an answer the client offers is an answer the
// resolver accepts.
//
// Three shapes, because CR 701.8a says a player discards as many as
// they can and no more:
//
//   - An empty hand discards nothing and the spell is done.
//   - A one-card hand discards that card, land or not. There is no
//     "two cards" to demand and the land clause buys nothing.
//   - Any larger hand gets the real choice: two cards, or one land.
//
// The hand is measured AFTER the draw, as printed ("then"), so the
// three fresh cards are among the ones that may be pitched.
//
// Targets a player, so it can be pointed at an opponent — rarely what
// you want, since they choose what to pitch, but it is the printed
// card and the group-hug decks that play it mean it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f947808c-a1cc-41ea-9d06-e1279a0da527",
		Name:         "Compulsive Research",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			who := item.Targets[0].ID
			if err := (DrawCards{Player: who, N: 3}).Apply(ctx); err != nil {
				return err
			}
			p := ctx.Game.PlayerByIDForEffect(who)
			if p == nil {
				return nil
			}
			switch held := len(p.Hand.Cards); {
			case held == 0:
				return nil
			case held == 1:
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   who,
					Source:   ctx.Source(),
					N:        1,
					Question: "Compulsive Research — discard your last card",
				})
			default:
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   who,
					Source:   ctx.Source(),
					N:        2,
					UpTo:     true,
					Question: "Compulsive Research — discard two cards, or one land card",
					Validate: func(picked []game.Card) bool {
						switch len(picked) {
						case 2:
							return true
						case 1:
							return picked[0].IsLand()
						default:
							return false
						}
					},
				})
			}
			return nil
		},
	})
}
