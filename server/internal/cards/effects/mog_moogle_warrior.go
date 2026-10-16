package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mog, Moogle Warrior — Legendary Creature — Moogle Warrior {1}{R}{W},
// 1/2 (deck request #3655):
//
//	"Lifelink
//	 Dance — At the beginning of your end step, each player may discard a
//	 card. Each player who discarded a card this way draws a card. If a
//	 creature card was discarded this way, you create a 1/2 white Moogle
//	 creature token with lifelink. Then if a noncreature card was
//	 discarded this way, put a +1/+1 counter on each Moogle you control."
//
// Each player, starting with the controller, chooses whether to discard
// and which card, then draws for it; the token and the counters come
// once the last player has answered, so "a creature card was discarded"
// and "a noncreature card was discarded" read the whole table. The new
// Moogle is one of the Moogles that get the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f18cce8c-8fb6-49a7-8f79-2529c0ee3ecc",
		Name:            "Mog, Moogle Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Mog, Moogle Warrior — each player may discard a card and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return mogAsk(ctx, tablePlayers(ctx), &mogDance{})
				}),
		},
	})
}

// mogDance is what the table discarded this resolution.
type mogDance struct{ creature, noncreature bool }

func mogAsk(ctx *Context, players []uuid.UUID, dance *mogDance) error {
	if len(players) == 0 {
		return mogFinish(ctx, dance)
	}
	player, rest := players[0], players[1:]
	item := ctx.Item
	return MayChoice{
		Player:   player,
		Question: "Mog, Moogle Warrior — discard a card (and draw a card)?",
		OnYes: func(c *Context) error {
			c.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   player,
				Source:   c.Source(),
				N:        1,
				Question: "Mog, Moogle Warrior — discard a card",
				Then: func(g *game.Game, seat uuid.UUID, discarded []uuid.UUID) error {
					for _, id := range discarded {
						if card, ok := g.LookupCardForEffect(id); ok {
							if card.IsCreature() {
								dance.creature = true
							} else {
								dance.noncreature = true
							}
						}
					}
					if len(discarded) > 0 {
						if err := g.DrawNForEffect(seat, 1); err != nil {
							return err
						}
					}
					return mogAsk(NewContext(g, item), rest, dance)
				},
			})
			return nil
		},
		OnNo: func(c *Context) error {
			return mogAsk(c, rest, dance)
		},
	}.Apply(ctx)
}

// mogFinish makes the Moogle if a creature card went, then counters the
// Moogles if a noncreature card went.
func mogFinish(ctx *Context, dance *mogDance) error {
	if dance.creature {
		if err := (CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/2 white Moogle with lifelink"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if !dance.noncreature {
		return nil
	}
	var moogles []uuid.UUID
	// "Moogle" is a word of the source's text (CR 612.2, #3984).
	moogle := ctx.TextSubtype("Moogle")
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.Controller == ctx.Controller() && c.HasSubtype(moogle) {
			moogles = append(moogles, c.InstanceID)
		}
	}
	for _, id := range moogles {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
