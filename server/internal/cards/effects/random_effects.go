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
	// TailCarry are IDs read as the effect resolves that the tail reads
	// (TailOutcome.Tail.Carry): the cards Danse Macabre's sacrifices put
	// into graveyards.
	TailCarry []uuid.UUID
	// Modifier is what the instruction adds to the roll, negative for
	// "subtract" (CR 706.2, #5681): "roll a d20 and add the number of
	// cards in your hand". Read it as the effect resolves and put it
	// here, never inside the tail: the roll is announced with its natural
	// result and this beside it, so "whenever you roll a natural 20"
	// reads the die and "if any of those results was 10 or higher" reads
	// the sum. The tail reads the sum as DiceResult(o).
	Modifier int
}

// Apply rolls the dice for the effect's controller.
func (r RollDice) Apply(ctx *Context) error {
	var then game.KeyedTail
	if r.Then.Key() != "" {
		then = withN(ctx.ItemTail(r.Then), r.TailN)
		then.Carry = append([]uuid.UUID(nil), r.TailCarry...)
	}
	d := randomDraw(ctx)
	return ctx.Game.RollDiceThenForEffect(game.DiceRoll{
		Player: d.Player, Source: d.Source, Sides: r.Sides, N: r.N, IgnoreLowest: r.IgnoreLowest, Then: then,
		Modifier: r.Modifier,
	})
}

// DiceResult is the result of the first kept die a dice tail was told:
// its natural result plus the instruction's modifier (CR 706.2,
// RollDice.Modifier). The number a results table reads (CR 706.3a). Zero
// when no die was kept.
func DiceResult(o game.TailOutcome) int {
	if len(o.Rolls) == 0 {
		return 0
	}
	return o.Rolls[0] + o.RollModifier
}

// RollOneMoreDieAndIgnoreTheLowest is the static "If you would roll one or
// more dice, instead roll that many dice plus one and ignore the lowest
// roll." (Pixie Guide, Barbarian Class; CR 614.1a, 706.6, #5681): a
// replacement on the dice roll's window (game.RepEventRollDice) for every
// roll its controller makes. One more die is rolled and one more of the
// lowest is ignored (ReplacementEvent.RollOneMoreAndIgnoreLowest); an
// ignored roll never happened, so nothing triggers on it. Two of them
// roll two more dice and ignore the two lowest. The planar die has no
// numerical result and so no lowest roll (CR 706.7): its roll is left
// alone. `name` is the card, for the label.
func RollOneMoreDieAndIgnoreTheLowest(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventRollDie},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventRollDice && !ev.RollPlanar && ev.RollCount > 0 && ev.RollPlayer == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.RollOneMoreAndIgnoreLowest()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:    name + ": roll one more die and ignore the lowest roll",
		RowLabel: "If you would roll one or more dice, instead roll that many dice plus one and ignore the lowest roll.",
	}
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

//nolint:unparam // n is the count "choose N at random" prints; every card so far picks one, in-flight batches may not.
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
//
// The effect is told the total of the results, each its natural result
// plus the instruction's modifier (CR 706.2, #5681). An effect that reads
// the dice one by one ("if any of those results was 10 or higher") calls
// RolledDice.
func WheneverYouRollDice(label string, effect func(*game.Game, *game.StackItem, int) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventRollDie}, Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Actor == source.Controller && ev.Seq == ev.BatchSeq
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			total := 0
			for _, rolled := range RolledDice(g, item) {
				total += rolled.RollResult()
			}
			return effect(g, item, total)
		},
	}
}

// RolledDice lists the dice of the instruction a "whenever you roll one or
// more dice" trigger fired on, one EventRollDie per kept die in the order
// rolled: Amount is a die's natural result, RollResult() its result with
// the instruction's modifier (CR 706.2). Ignored rolls are not there (CR
// 706.6), and neither is a planar die: it triggers the ability but has no
// numerical result (CR 706.7, 901.9d; ADR 0216 §4). The batch is read off
// the item's triggering event, so it is the same on a restored table.
func RolledDice(g *game.Game, item *game.StackItem) []game.Event {
	if item == nil || item.Trigger == nil {
		return nil
	}
	batch := item.Trigger.Event.BatchSeq
	var out []game.Event
	for _, rolled := range g.Events {
		if rolled.Kind == game.EventRollDie && rolled.BatchSeq == batch && rolled.PlanarFace == "" {
			out = append(out, rolled)
		}
	}
	return out
}

// YouRolledADiesHighestNaturalResult is "whenever you roll a die's highest
// natural result" (Netherese Puzzle-Ward, #5681): one EventRollDie of the
// source's controller whose die shows its highest face. The natural
// result is the number on the die before any modifier (CR 706.2), so the
// event's Amount, never its RollResult. One event per kept die, so two
// dice that both show their highest face fire the ability twice; an
// ignored roll never happened (CR 706.6) and a planar die has no number.
func YouRolledADiesHighestNaturalResult(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventRollDie && ev.PlanarFace == "" && ev.Actor != uuid.Nil && ev.Actor == source.Controller &&
		ev.Sides > 0 && ev.Amount == ev.Sides
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
