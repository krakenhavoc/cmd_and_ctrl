package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b153_helpers.go — GAP batch 153 (Jumpstart 2022 green creatures,
// #5948). Each name is the card's, so a parallel batch cannot collide.

// gapB153ThreeOrMorePlusOneCounters is "creatures you control with three or
// more +1/+1 counters on them" (Runadi, Behemoth Caller), read live off the
// counter map: a counter is a marker the layers read, not a layered value.
func gapB153ThreeOrMorePlusOneCounters(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		target.Counters[game.CounterPlusOne] >= 3
}

// gapB153RunadiCastTrigger is "Whenever you cast a creature spell with mana
// value 5 or greater, that creature enters with X additional +1/+1 counters
// on it, where X is its mana value minus 4." X is read as the spell is cast
// (an {X} spell counts the X it was cast with, CR 202.3e) and carried on the
// trigger; on resolution the spell is marked so the counters ride its entry,
// as Wildgrowth Archaic's do. A spell gone by resolution has nothing to mark.
func gapB153RunadiCastTrigger(label string) game.TriggeredAbility {
	t := On(game.EventCast, YouCast(And(Creature(), ManaValueGE(5))), label, func(g *game.Game, item *game.StackItem) error {
		n := item.Params.Amount
		if n <= 0 {
			return nil
		}
		g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
			Kind:       game.CounterPlusOne,
			Count:      n,
			Source:     item.SourceCardID,
			SourceName: "Runadi, Behemoth Caller",
		})
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
			mv, _ := g.ManaValueForEffect(spell)
			item.Params.Amount = mv - 4
		}
		return item
	}
	return t
}

// gapB153BandarMove is Scrounging Bandar's upkeep trigger: "you may move any
// number of +1/+1 counters from this creature onto another target creature".
// The target is chosen as the trigger goes on the stack; how many move is
// chosen as it resolves (CR 608.2c), zero being the "may" declined.
func gapB153BandarMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if !onBattlefield(g, self) || sourceIsNewObject(g, item) {
		return nil
	}
	to, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	have := 0
	if c, ok := g.LookupCardForEffect(self); ok {
		have = c.Counters[game.CounterPlusOne]
	}
	if have <= 0 {
		return nil
	}
	return ChooseNumber{
		Min:      0,
		Max:      have,
		Question: "Scrounging Bandar — how many +1/+1 counters to move onto the target creature?",
		Unit:     game.PayAmountCounters,
		Then:     gapB153BandarMoved,
		Carry:    []uuid.UUID{self, to},
	}.Apply(ctx)
}

// gapB153BandarMoved moves the chosen number: off the Bandar, onto the
// target, only while both are still on the battlefield (CR 122.5). An on-disk
// key: never renamed, never reused.
var gapB153BandarMoved = NumberThen("number/s82-gap-b153-scrounging-bandar-move",
	func(ctx *Context, r game.NumberAnswer) error {
		if len(r.Carry) < 2 || r.Number <= 0 {
			return nil
		}
		self, to := r.Carry[0], r.Carry[1]
		if !onBattlefield(ctx.Game, self) || !onBattlefield(ctx.Game, to) {
			return nil
		}
		moved, err := socB9RemoveCounters(ctx, self, game.CounterPlusOne, r.Number)
		if err != nil || moved <= 0 {
			return err
		}
		return AddCounter{Target: to, Kind: game.CounterPlusOne, N: moved}.Apply(ctx)
	})

// gapB153CreatureWithCounterTargetedBySpell is Swarm Shambler's watch:
// a creature its controller controls with a +1/+1 counter on it became the
// target of a SPELL an opponent controls. EventBecomesTarget fires once per
// target slot (CR 115.3), so only the first event of one spell for one
// creature counts.
func gapB153CreatureWithCounterTargetedBySpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.TargeterKind != game.StackItemSpell {
		return false
	}
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !onBattlefield(g, ev.CardID) || !c.IsCreature() || c.Controller != source.Controller ||
		c.Counters[game.CounterPlusOne] <= 0 {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		earlier := g.Events[i]
		if earlier.Seq >= ev.Seq {
			continue
		}
		switch earlier.Kind {
		case game.EventBecomesTarget:
			if earlier.StackItemID == ev.StackItemID && earlier.CardID == ev.CardID {
				return false
			}
		case game.EventTrigger, game.EventTriggerPlaced:
			continue
		default:
			return true
		}
	}
	return true
}

// gapB153GreatestManaValueAmongYourCreatures is the greatest mana value among
// the creatures `controller` controls (Towering Gibbon), 0 with none.
func gapB153GreatestManaValueAmongYourCreatures(g *game.Game, controller uuid.UUID) int {
	greatest := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if mv, ok := g.ManaValueForEffect(*c); ok && mv > greatest {
			greatest = mv
		}
	}
	return greatest
}

// gapB153WickedWolfSacrificeFood is "Put a +1/+1 counter on this creature. It
// gains indestructible until end of turn. Tap it." Each part acts only while
// the Wolf is still the permanent that activated the ability; the counter and
// the keyword land first, as printed, so the tap that follows does not undo
// the protection.
func gapB153WickedWolfSacrificeFood(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := plusOneCountersOnThis(1)(g, item); err != nil {
		return err
	}
	if err := thisGainsKeywordUntilEndOfTurn("indestructible", "Wicked Wolf — indestructible until end of turn")(g, item); err != nil {
		return err
	}
	return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
