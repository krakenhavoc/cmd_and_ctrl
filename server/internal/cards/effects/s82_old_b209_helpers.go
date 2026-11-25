package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b209_helpers.go — shared bodies for Dominaria counters, batch 209 (#5530).

// OldB209YouCastAKickedSpell is "whenever you cast a spell, if that spell was
// kicked": your cast, of a spell whose payment record shows a kicker paid (once
// or more), read off its stack item.
func OldB209YouCastAKickedSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.KickedTimesPaid(spell, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
}

// OldB209HallarKicked is Hallar's body: a +1/+1 counter on the source while it
// is still that permanent, then damage equal to its +1/+1 counters to each
// opponent as one instance (counters as they were if it has gone, CR 608.2h).
func OldB209HallarKicked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if SourceIsStillThisPermanent(g, item) {
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	n := CountersOnThisAsItResolves(game.CounterPlusOne)(ctx)
	if n <= 0 {
		return nil
	}
	return DamageToEachOpponent(g, item, n)
}

// OldB209OpponentCreatureDealtDamage is "a creature an opponent controls is dealt
// damage": a creature on the battlefield, controlled by someone other than the
// source's controller, was dealt damage (B28CreatureWasDealtDamage).
func OldB209OpponentCreatureDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := B28CreatureWasDealtDamage(ev, g)
	return ok && c.Controller != source.Controller
}

// OldB209LoreCounterOnTarget puts a lore counter on the first still-legal target.
func OldB209LoreCounterOnTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterLore, N: 1}.Apply(ctx)
}

// OldB209EggCounter is the counter Darigaaz Reincarnated is exiled with.
const OldB209EggCounter = "egg"

// OldB209DarigaazDies is "if Darigaaz would die, instead exile it with three egg
// counters on it": a replacement of the permanent's own move from the
// battlefield to a graveyard.
func OldB209DarigaazDies() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && src != nil && ev.CardID == src.InstanceID &&
				ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			ev.ExileCounter = OldB209EggCounter
			ev.ExileCounterCount = 3
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Darigaaz Reincarnated: exile it with three egg counters instead of dying",
	}
}

// OldB209DarigaazUpkeep removes an egg counter from the exiled card and, when
// none is left, returns it to the battlefield under its owner's control. It
// rechecks "exiled with an egg counter" first (CR 603.4).
func OldB209DarigaazUpkeep(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if g.Exile == nil || !g.Exile.Contains(id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[OldB209EggCounter] <= 0 {
		return nil
	}
	// The card's counters are one map the removal edits in place, so the
	// count left is read before it.
	left := c.Counters[OldB209EggCounter] - 1
	if err := g.AddCounterForEffect(id, OldB209EggCounter, -1); err != nil {
		return err
	}
	if left > 0 {
		return nil
	}
	_, err := g.ReturnFromExileToBattlefieldForEffect(id, c.Owner, false)
	return err
}

// OldB209ChooseAimTargets is Haphazard Bombardment's enters body: its controller
// chooses four nonenchantment permanents they don't control (all of them when
// there are fewer), none targeted, and an aim counter goes on each.
func OldB209ChooseAimTargets(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller && !c.HasCardType("Enchantment") {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if len(ids) <= 4 {
		return oldB209PutAim(g, item.SourceCardID, ids)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Haphazard Bombardment — choose four nonenchantment permanents you don't control",
		Cards:    ids,
		Min:      4,
		Max:      4,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  oldB209AimPicked,
	})
	return nil
}

// OldB209AimCounter is the counter Haphazard Bombardment puts on its choices.
const OldB209AimCounter = "aim"

var oldB209AimPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b209-aim-counters", func(g *game.Game, r game.CardsChosen) error {
	return oldB209PutAim(g, r.Source, r.Picked)
})

func oldB209PutAim(g *game.Game, _ uuid.UUID, ids []uuid.UUID) error {
	for _, id := range ids {
		if !g.Battlefield.Contains(id) {
			continue
		}
		if err := g.AddCounterForEffect(id, OldB209AimCounter, 1); err != nil {
			return err
		}
	}
	return nil
}

// OldB209AimedPermanents lists the permanents `you` don't control that have an
// aim counter on them.
func OldB209AimedPermanents(g *game.Game, you uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.Counters[OldB209AimCounter] > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// OldB209DestroyOneAimedAtRandom destroys one permanent its controller doesn't
// control that has an aim counter, chosen at random, if there are still two or
// more (CR 603.4).
func OldB209DestroyOneAimedAtRandom(g *game.Game, item *game.StackItem) error {
	ids := OldB209AimedPermanents(g, item.Controller)
	if len(ids) < 2 {
		return nil
	}
	pick := g.ChooseAtRandomForEffect(game.RandomDraw{Player: item.Controller, Source: item.SourceCardID}, ids, 1)
	if len(pick) == 0 {
		return nil
	}
	return DestroyTarget{Target: pick[0]}.Apply(NewContext(g, item))
}
