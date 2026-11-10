package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b26_helpers.go — the shared pieces of Standard-set batch 26
// (Duskmourn: House of Horror, #4466; Sprint S79, #4452). One file for
// the batch so a parallel batch never edits the same one.

// --- Undead Sprinter ---------------------------------------------------

// stdB26NonZombieCreatureDied is "a non-Zombie creature died this turn",
// asked of the card's owner when the cast is offered. The table-wide count
// of creatures that went to a graveyard this turn, less those that were
// Zombies as they died (DiedWithSubtypeThisTurn reads the types the
// creature had on the battlefield, a changeling included, CR 603.10a).
// Tokens count: each died as a creature before it ceased to exist.
func stdB26NonZombieCreatureDied(g *game.Game, _ uuid.UUID, card game.Card) bool {
	// The word is read through the card whose text it is (CR 612.2).
	zombie := card.TextSubtype("Zombie")
	zombies := 0
	for _, p := range g.Seats {
		if p != nil {
			zombies += g.DiedWithSubtypeThisTurn(p.ID, zombie)
		}
	}
	return g.TurnTally.CreaturesDied > zombies
}

// --- Attack-in-the-Box -------------------------------------------------

// stdB26AttackInTheBox is the "you may have it get +4/+0 until end of
// turn. If you do, sacrifice it at the beginning of the next end step"
// trigger's resolution, run once the player has said yes. A box that has
// left, or come back as a new object, is not "it" any more (CR 400.7).
func stdB26AttackInTheBox(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 4, Label: "Attack-in-the-Box +4/+0"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: "Attack-in-the-Box — sacrifice it",
		Cards: []uuid.UUID{item.SourceCardID},
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}

// --- Come Back Wrong ---------------------------------------------------

// stdB26ComeBackWrongDestroyed is "If a creature card is put into a
// graveyard this way, return it to the battlefield under your control.
// Sacrifice it at the beginning of your next end step." It is told the
// permanents that were destroyed and landed in a graveyard; a token is not
// a card, and a card that was sent elsewhere (a commander to the command
// zone) is not "put into a graveyard this way". Nothing can happen between
// the destruction and the return, so this runs inside the resolution.
var stdB26ComeBackWrongDestroyed = CardTail("std-b26/come-back-wrong-return", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsToken() || !c.IsCreature() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (ScheduleDelayedTrigger{
			At:                 game.StepEnd,
			Label:              "Come Back Wrong — sacrifice it",
			Controller:         ctx.Controller(),
			ControllerTurnOnly: true,
			Cards:              []uuid.UUID{id},
			Body:               sacrificeListedCardsBody,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// --- Cursed Recording --------------------------------------------------

// stdB26CursedRecordingCounter is the time counter the artifact collects.
const stdB26CursedRecordingCounter = "time"

// stdB26CursedRecordingTick is the "Then if there are seven or more time
// counters on it, remove those counters and it deals 20 damage to you"
// half of the trigger. The count is read as the ability resolves, after
// the counter has landed; an artifact that left is not "it".
func stdB26CursedRecordingTick(ctx *Context) error {
	if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || c.Counters[stdB26CursedRecordingCounter] < 7 {
		return nil
	}
	n := c.Counters[stdB26CursedRecordingCounter]
	if err := (AddCounter{Target: ctx.Source(), Kind: stdB26CursedRecordingCounter, N: -n}).Apply(ctx); err != nil {
		return err
	}
	return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 20}.Apply(ctx)
}

// --- Ghost Vacuum ------------------------------------------------------

// stdB26GhostVacuumExile is "{T}: Exile target card from a graveyard", the
// exile linked to this artifact (CR 607.2a).
func stdB26GhostVacuumExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return ExileTarget{Target: id, Linked: true}.Apply(ctx)
}

// stdB26GhostVacuumReanimate is "Put each creature card exiled with this
// artifact onto the battlefield under your control with a flying counter on
// it. Each of them is a 1/1 Spirit in addition to its other types." The
// artifact was sacrificed to pay for it, so the link is read off the object
// the ability came from, not the permanent (CR 607.2a, CR 608.2h). All of
// them land as one entry.
func stdB26GhostVacuumReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return putCardsFromZoneTogetherWith(ctx, game.ZoneExile, ids, game.ZoneEntryOptions{
		Controller:         item.Controller,
		EntersWithCounters: map[string]int{game.CounterFlying: 1},
		EntersAs: &game.EntersAs{
			AddTypes:    []string{"Creature"}, // already creature cards; the overwrite names a type to add
			AddSubtypes: ctx.TextSubtypes("Spirit"),
			BasePT:      &game.EntersAsPT{Power: 1, Toughness: 1},
			Label:       "Ghost Vacuum — a 1/1 Spirit in addition to its other types",
		},
	}, game.TailRef{}, 0)
}

// --- Niko, Light of Hope -----------------------------------------------

// stdB26NikoExile is the activated ability's resolution: exile the target,
// and the rest, which waits for the exile, is the tail. The creature's
// object epoch rides the tail, because the object is a different one once
// it sits in exile and its copiable values are read by the object it was.
func stdB26NikoExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ExileTarget{
		Target:   id,
		ThenTail: withN(withCarry(ctx.ItemTail(stdB26NikoExiled), id), c.ObjectEpoch),
	}.Apply(ctx)
}

// stdB26NikoExiled is "Shards you control become copies of it until the
// next end step. Return it to the battlefield under its owner's control at
// the beginning of the next end step." Each Shard the controller has now is
// a copy of the creature as it last existed on the battlefield, and the
// creature, if it really reached exile, comes back at the next end step
// under its owner's control (the blink delayed trigger).
var stdB26NikoExiled = CardTail("std-b26/niko-after-exile", func(ctx *Context, o game.TailOutcome) error {
	id := tailCarry0(o)
	if id == uuid.Nil {
		return nil
	}
	var shards []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype("Shard") {
			shards = append(shards, c.InstanceID)
		}
	}
	if info, ok := ctx.Game.PermanentForEffect(game.ObjectRef{ID: id, Epoch: o.Tail.N}); ok && info.Copiable != nil && len(shards) > 0 {
		if err := (BecomeCopy{
			Targets:  shards,
			Values:   info.Copiable,
			Duration: CopyUntilNextEndStep,
			Label:    "Niko, Light of Hope — a copy of the exiled creature until the next end step",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: "Niko, Light of Hope — return the exiled creature",
		Cards: []uuid.UUID{id},
		Body:  returnExiledToOwnersBody,
	}.Apply(ctx)
})

// --- Reluctant Role Model ----------------------------------------------

// stdB26RoleModelKinds are the counters the Survival ability offers, in
// printed order: "a flying, lifelink, or +1/+1 counter".
var stdB26RoleModelKinds = []string{game.CounterFlying, game.CounterLifelink, game.CounterPlusOne}

// stdB26RoleModelPick is the Survival ability's effect: the controller
// picks one of the three counters as it resolves. A creature that left, or
// came back as a new object, gets nothing (CR 400.7); the "if it's tapped"
// half is Survival's own.
func stdB26RoleModelPick(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return PickOption{
		Question: "Reluctant Role Model — put a flying, lifelink, or +1/+1 counter on it",
		Options: []game.ChoiceOption{
			{Label: "A flying counter"}, {Label: "A lifelink counter"}, {Label: "A +1/+1 counter"},
		},
		ThenKey:  stdB26RoleModelPicked,
		WithItem: true,
	}.Apply(NewContext(g, item))
}

// stdB26RoleModelPicked puts the chosen counter on the Role Model, if it is
// still the permanent the ability came from. An on-disk key.
var stdB26RoleModelPicked = OptionPickThen("option-pick/std-b26-reluctant-role-model-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(stdB26RoleModelKinds) || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return ctx.Game.AddCounterByForEffect(ctx.Controller(), ctx.Source(), stdB26RoleModelKinds[r.Index], 1)
})

// stdB26RoleModelCreatureDied is the trigger condition of "Whenever this
// creature or another creature you control dies, if it had counters on it":
// a creature that died under your control (the source included, CR 603.10a)
// whose last-known information holds a counter (the intervening if, CR 603.4).
func stdB26RoleModelCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	for _, n := range g.LastKnownCountersForEffect(ev.CardID) {
		if n > 0 {
			return true
		}
	}
	return false
}

// stdB26RoleModelMoveCounters is "put those counters on up to one target
// creature": the same number of each kind the creature had as it died
// (its last-known counters), put on by the ability, so a replacement that
// changes counters applies to it (the card's rulings). No target, or an
// illegal one, does nothing.
func stdB26RoleModelMoveCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := firstLegalCardTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	left, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	kinds := make([]string, 0, len(left.Counters))
	for kind, n := range left.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if err := g.AddCounterByForEffect(item.Controller, target, kind, left.Counters[kind]); err != nil {
			return err
		}
	}
	return nil
}

// --- Nashi, Searcher in the Dark ---------------------------------------

// stdB26NashiMill is the trigger's resolution: mill that many cards. The
// rest of the instruction is the tail, told the cards that landed.
func stdB26NashiMill(g *game.Game, item *game.StackItem) error {
	n := item.Params.Amount
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	return MillToZone{N: n, ThenTail: ctx.ItemTail(stdB26NashiMilled)}.Apply(ctx)
}

// stdB26NashiMilled is "You may put any number of legendary and/or
// enchantment cards from among them into your hand. If you put no cards
// into your hand this way, put a +1/+1 counter on Nashi." With nothing to
// offer no question is asked and the counter goes on.
var stdB26NashiMilled = CardTail("std-b26/nashi-after-mill", func(ctx *Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !(c.IsLegendary() || c.HasCardType("enchantment")) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return stdB26NashiCounter(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Nashi, Searcher in the Dark — put any number of legendary and/or enchantment cards from among them into your hand",
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  stdB26NashiPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// stdB26NashiPicked moves the chosen cards to the hand and, when none of
// them got there, puts the counter on. An on-disk key.
var stdB26NashiPicked = askedPick("choose-cards/std-b26-nashi-to-hand", func(ctx *Context, r game.CardsChosen) error {
	taken := 0
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			taken++
		}
	}
	if taken > 0 {
		return nil
	}
	return stdB26NashiCounter(ctx)
})

// stdB26NashiCounter is the +1/+1 counter on Nashi, if it is still the
// permanent the ability came from.
func stdB26NashiCounter(ctx *Context) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
