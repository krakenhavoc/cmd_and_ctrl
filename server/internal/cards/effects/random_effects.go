package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// randomDraw identifies the player and object whose effect consumes a keyed
// random stream. It deliberately contains IDs only so every continuation is
// safe to clone and restore.
func randomDraw(ctx *Context) game.RandomDraw {
	return game.RandomDraw{Player: ctx.Controller(), Source: ctx.Source()}
}

func rollDice(ctx *Context, sides, n int) ([]int, error) {
	return ctx.Game.RollDiceForEffect(randomDraw(ctx), sides, n)
}

func randomPick(ctx *Context, ids []uuid.UUID, n int) []uuid.UUID {
	return ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), ids, n)
}

func graveyardIDs(g *game.Game, player uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		if match == nil || match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func treasureTokens(g *game.Game, controller uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return g.CreateTokenForEffect(controller, TreasureToken(), n)
}

func randomCardToHandOrBattlefield(g *game.Game, id uuid.UUID) error {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if c.IsCreature() {
		return g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield)
	}
	return g.ReturnFromGraveyardForEffect(id, game.ZoneHand)
}

// WheneverYouRollDice fires once per instruction. Unlike an in-flight guard,
// BatchSeq distinguishes a second instruction even while the first trigger
// is still waiting on the stack. The batch is read off the item's
// triggering event (item.Trigger.Event.BatchSeq, ADR 0041 P9) rather than
// captured, so a table with this trigger waiting on the stack is a
// restore point.
func WheneverYouRollDice(label string, effect func(*game.Game, *game.StackItem, int) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventRollDie}, Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Actor == source.Controller && ev.Seq == ev.BatchSeq
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			batch := item.Trigger.Event.BatchSeq
			total := 0
			for _, rolled := range g.Events {
				if rolled.Kind == game.EventRollDie && rolled.BatchSeq == batch {
					total += rolled.Amount
				}
			}
			return effect(g, item, total)
		},
	}
}
