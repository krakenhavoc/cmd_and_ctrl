package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b23_helpers.go — helpers for the +1/+1 counter commander cards of
// cmdr2 batch 23 (#5368).

// cmdr2B23HasPlusOneCounter reports whether the card `id` is on the
// battlefield with at least one +1/+1 counter, read now.
func cmdr2B23HasPlusOneCounter(g *game.Game, id uuid.UUID) bool {
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Counters[game.CounterPlusOne] > 0
}

// cmdr2B23OtherWizardsEnterWithACounter is Sage of Fables' CR 614
// entry replacement: "each other Wizard creature you control enters
// with an additional +1/+1 counter on it" (Grumgully's shape).
func cmdr2B23OtherWizardsEnterWithACounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller && entering.HasSubtype("Wizard")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// cmdr2B23AnotherCreatureWithPlusOne is "you control another creature
// with a +1/+1 counter on it": a creature `you` control other than `self`.
func cmdr2B23AnotherCreatureWithPlusOne(g *game.Game, you, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.IsCreature() && c.Controller == you && c.Counters[game.CounterPlusOne] > 0 {
			return true
		}
	}
	return false
}

// cmdr2B23CreatureYouControlWithPlusOne is the CardPredicate for "a
// creature you control with a +1/+1 counter on it", read live.
func cmdr2B23CreatureYouControlWithPlusOne(_ *game.Game, you uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Controller == you && c.Counters[game.CounterPlusOne] > 0
}

// cmdr2B23GrowthCounter is Simic Ascendancy's counter kind.
const cmdr2B23GrowthCounter = "growth"

// cmdr2B23DoubleThisPower is "double this creature's power until end of
// turn" (CR 701.10b): +X/+0 where X is the source's power as the ability
// resolves, read unclamped (Casey Jones, Asphalt Hooligan). Nothing when
// the source is gone or is a new object (CR 400.7).
func cmdr2B23DoubleThisPower(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !sourceIsStillThisPermanent(g, item) {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: c.PowerForComparison(), Label: label}.Apply(NewContext(g, item))
	}
}

// cmdr2B23ShuffleTheRestTail is "then shuffle the rest into your
// library" after a reveal-and-put: the revealed cards that were not put
// are still in the library, so the rest is shuffled with it. An on-disk
// key: never renamed, never reused.
var cmdr2B23ShuffleTheRestTail = PutFromLibraryTail("put-from-library/cmdr2-shuffle-the-rest",
	func(ctx *Context, res PutFromLibraryResult) error {
		return ctx.Game.ShuffleLibraryForEffect(res.Player)
	})

// cmdr2B23MarauderMoves is Aetherborn Marauder's clause: "move any number of
// +1/+1 counters from other permanents you control onto this creature". The
// chain asks once per source permanent, in battlefield order, how many
// +1/+1 counters to move (zero is allowed); the Marauder and the sources
// still to be asked ride the prompt as the clause's extra IDs, so a table
// waiting on any question is a restore point.
var cmdr2B23MarauderMoves *counterChoice

// The clause is registered in init, not as a package variable: its final step
// asks the next question through the clause itself, which a variable
// initialiser cannot refer back to.
func init() {
	cmdr2B23MarauderMoves = registerCounterChoice("number/cmdr2-aetherborn-marauder-move", nil, cmdr2B23MarauderFinal)
}

func cmdr2B23MarauderFinal(ctx *Context, from uuid.UUID, extra []uuid.UUID, picked map[string]int) error {
	if len(extra) == 0 {
		return nil
	}
	marauder, rest := extra[0], extra[1:]
	if n := picked[game.CounterPlusOne]; n > 0 && onBattlefield(ctx.Game, marauder) && onBattlefield(ctx.Game, from) {
		if src, ok := ctx.Game.LookupCardForEffect(from); ok {
			n = min(n, src.Counters[game.CounterPlusOne])
			if n > 0 {
				if err := (AddCounter{Target: from, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: marauder, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
		}
	}
	return cmdr2B23MarauderAsk(ctx, marauder, rest)
}

// cmdr2B23MarauderAsk asks about the first of `sources` that still has +1/+1
// counters and is still a permanent its controller controls, then carries the
// rest on to the next question.
func cmdr2B23MarauderAsk(ctx *Context, marauder uuid.UUID, sources []uuid.UUID) error {
	if !onBattlefield(ctx.Game, marauder) {
		return nil
	}
	for i, id := range sources {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) || c.Controller != ctx.Controller() || c.Counters[game.CounterPlusOne] <= 0 {
			continue
		}
		extra := append([]uuid.UUID{marauder}, sources[i+1:]...)
		return cmdr2B23MarauderMoves.ask(ctx, counterAsk{
			perm: id, extra: extra, question: "Aetherborn Marauder — move +1/+1 counters from " + c.Name,
			kinds: []string{game.CounterPlusOne}, have: []int{c.Counters[game.CounterPlusOne]}, picked: []int{0},
		})
	}
	return nil
}

// cmdr2B23MarauderMove is the enters trigger's body: the sources are every
// other permanent the controller controls with a +1/+1 counter as it resolves.
func cmdr2B23MarauderMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var sources []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != item.SourceCardID && c.Controller == ctx.Controller() && c.Counters[game.CounterPlusOne] > 0 {
			sources = append(sources, c.InstanceID)
		}
	}
	return cmdr2B23MarauderAsk(ctx, item.SourceCardID, sources)
}

// cmdr2B23IngredientCounter is Communal Brewing's counter kind.
const cmdr2B23IngredientCounter = "ingredient"

// cmdr2B23Graft is the triggered half of "Graft N" (CR 702.58a), written once
// for the only graft card in the catalog: "Whenever another creature enters,
// if this permanent has a +1/+1 counter on it, you may move a +1/+1 counter
// from this permanent onto that creature". Any creature counts, whoever
// controls it, and the intervening "if" (CR 603.4) is checked as the creature
// enters and again as the trigger resolves. The move needs both permanents
// still on the battlefield. The static half, "this permanent enters with N
// +1/+1 counters on it", is the card's own entry replacement (a CR 614.1c
// clause), which a card that also enters tapped must fold into one
// replacement so the table is not asked to order two (SelfEntersTappedWithCounters).
func cmdr2B23Graft(name string) game.TriggeredAbility {
	label := name + " — graft: you may move a +1/+1 counter from this permanent onto that creature"
	return InterveningIf(CondBoard{
		Name: "this permanent has a +1/+1 counter on it",
		Test: func(ctx *Context) bool {
			return cmdr2B23HasPlusOneCounter(ctx.Game, ctx.Item.SourceCardID)
		},
	}, Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventETB || source == nil || ev.CardID == source.InstanceID || source.Counters[game.CounterPlusOne] <= 0 {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature()
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || sourceIsNewObject(g, item) || !cmdr2B23HasPlusOneCounter(g, item.SourceCardID) {
			return nil
		}
		to := item.Trigger.Event.CardID
		if !onBattlefield(g, to) {
			return nil
		}
		ctx := NewContext(g, item)
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: -1}).Apply(ctx); err != nil {
			return err
		}
		return AddCounter{Target: to, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}), name+" — move a +1/+1 counter onto the creature that entered?"))
}
