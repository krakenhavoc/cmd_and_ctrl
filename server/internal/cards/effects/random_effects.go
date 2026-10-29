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

// RollDice is "roll N dSides" with the rest of the effect after it: every
// card's roll goes through the engine's one dice door
// (game.RollDiceThenForEffect, #4307), so a replacement of the roll
// (Squid Fire Knight's sticker: one more die, then its controller chooses
// one to ignore) applies, and the rest of the effect waits for that
// answer. Then is a registered tail (DiceTail), told the kept rolls in
// the order rolled; it runs against the resolving item, carried as data,
// so a table waiting on the dice question is a restore point.
// IgnoreLowest is the card's own "ignore the lowest" (CR 706.6,
// Berserker's Frenzy), applied after any roll a player chose to ignore.
type RollDice struct {
	Sides, N, IgnoreLowest int
	Then                   game.TailRef
	// TailN is a number read as the effect resolves that the tail reads
	// (TailOutcome.Tail.N): Diviner's Portent's X.
	TailN int
}

// Apply rolls the dice for the effect's controller.
func (r RollDice) Apply(ctx *Context) error {
	var then game.KeyedTail
	if r.Then.Key() != "" {
		then = withN(ctx.ItemTail(r.Then), r.TailN)
	}
	d := randomDraw(ctx)
	return ctx.Game.RollDiceThenForEffect(game.DiceRoll{
		Player: d.Player, Source: d.Source, Sides: r.Sides, N: r.N, IgnoreLowest: r.IgnoreLowest, Then: then,
	})
}

// DiceTail registers the rest of a dice-rolling effect, told the kept
// rolls (never empty when a die was rolled). The key is an on-disk
// identity (ledgered "tail <key>"): never renamed, never reused. Call it
// once, from a package-level var.
func DiceTail(key string, body func(ctx *Context, rolls []int) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if len(o.Rolls) == 0 {
			return nil
		}
		return body(ctx, o.Rolls)
	})
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
	return g.CreateTokenForEffect(controller, ruleDefinedToken(TreasureToken()), n)
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

// YouWonACoinFlip is "whenever you win a coin flip" (CR 705.2, ADR 0054
// Decision 6, amended 2026-10-15 for #3767): an EventFlipCoin that the
// source's controller flipped and won. Only the player who flips a coin
// wins or loses it, so the flipper is the event's Actor and nobody else's
// ability of this kind sees the flip. One event is emitted per coin, so a
// flip of two coins that wins both fires twice: once per won flip.
//
// A face-only flip (Game.FlipCoinsForEffect: nobody calls it, so nobody
// wins) carries Won false and never matches, and a coin flipped at the
// table for fun is an EventTableRoll, never this kind.
func YouWonACoinFlip(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventFlipCoin && ev.Won && ev.Actor != uuid.Nil && ev.Actor == source.Controller
}

// WheneverYouWinACoinFlip is "Whenever you win a coin flip, <effect>"
// (Chance Encounter, Tavern Scoundrel, Setzer, Wandering Gambler).
func WheneverYouWinACoinFlip(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventFlipCoin, YouWonACoinFlip, label, effect)
}

// YouLostACoinFlip is "whenever you lose a coin flip" (CR 705.2, #4063,
// Karplusan Minotaur): an EventFlipCoin that the source's controller
// flipped, called and did not win. The event's Call is what tells a lost
// flip from a face-only one: only a called flip (FlipCoinForEffect) has
// one, because only a called flip has anything to lose, while a flip of
// faces nobody wins or loses (FlipCoinsForEffect) carries Won false and no
// Call and never matches. One event per coin, so two lost coins fire the
// ability twice. A coin flipped at the table is an EventTableRoll, never
// this kind, and an opponent's lost flip is not yours (the flipper is the
// event's Actor).
func YouLostACoinFlip(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventFlipCoin && !ev.Won && ev.Call != "" && ev.Actor != uuid.Nil && ev.Actor == source.Controller
}

// WheneverYouLoseACoinFlip is "Whenever you lose a coin flip, <effect>"
// (Karplusan Minotaur).
func WheneverYouLoseACoinFlip(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventFlipCoin, YouLostACoinFlip, label, effect)
}

// FlipACoin is a bare "flip a coin" that prints nothing for winning or
// losing (Tavern Scoundrel's activated ability, Setzer's Vehicle trigger).
// The flip still has a winner: CR 705.2 has the flipper call it when
// something cares whether they win, and on both cards their own "whenever
// you win a coin flip" does. So it is Game.FlipCoinForEffect's called flip
// with no continuation, and what winning does is whatever triggers on the
// EventFlipCoin it emits.
//
// The flipper is the controller of the resolving ability.
type FlipACoin struct {
	// Question is the coin-call prompt's text.
	Question string
}

// Apply queues the coin call.
func (f FlipACoin) Apply(ctx *Context) error {
	flipper := ctx.Controller()
	if flipper == uuid.Nil {
		return nil
	}
	ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  flipper,
		Source:   ctx.Source(),
		Question: f.Question,
	})
	return nil
}
