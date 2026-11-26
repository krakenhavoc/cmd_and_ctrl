package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b321_helpers.go — helpers for Magic Origins, batch 321 (S82 older
// sets, #5549).

// OldB321ThisWouldDieExile is "If this creature would die, exile it instead"
// (Possessed Skaab, Wildwood Escort's replacement): a move of the source from
// the battlefield to a graveyard becomes a move to exile (CR 614.1a, CR 700.4).
func OldB321ThisWouldDieExile(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventMove && ev.CardID == src.InstanceID &&
				ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:    label + ": exile instead of dying",
		RowLabel: "If this creature would die, exile it instead.",
	}
}

// OldB321ErebosReturn is the rest of Erebos's Titan's trigger, "If you do,
// return this card from your graveyard to your hand": it runs once the may
// discard is settled, and only a card really discarded brings the Titan back,
// and only if it is still in the graveyard (CR 608.2b). An on-disk key: never
// renamed, never reused.
var OldB321ErebosReturn = game.RegisterDiscardThenRule("choose-cards/s82-old-b321-erebos-titan-return", nil,
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		if z := g.FindCardZoneForEffect(r.Source); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: r.Source, Dest: game.ZoneHand}.Apply(NewContext(g, r.Item))
	})

// OldB321ErebosDiscardReturn is "you may discard a card. If you do, return
// this card from your graveyard to your hand".
func OldB321ErebosDiscardReturn(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Erebos's Titan — discard a card to return it to your hand?",
	}, OldB321ErebosReturn, nil)
}

// OldB321CreatureCardLeftAnOpponentsGraveyard is "a creature card leaves an
// opponent's graveyard", asked of the card as it was in the graveyard (a spell
// cast from it can be another face): the card's owner is an opponent of the
// source's controller (CR 603.10a).
func OldB321CreatureCardLeftAnOpponentsGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil {
		return false
	}
	c, ok := CardThatLeftAGraveyard(ev, g)
	return ok && c.Owner != source.Controller && WasCardInAGraveyard(c, game.Card.IsCreature)
}

// OldB321SelfWhileOpponentsControlNoCreatures reads "as long as your opponents control
// no creatures" for the source alone.
func OldB321SelfWhileOpponentsControlNoCreatures(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.InstanceID != source.InstanceID {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != source.Controller && c.IsCreature() {
			return false
		}
	}
	return true
}

// OldB321InfiniteObliteration is "Choose a creature card name. Search target
// opponent's graveyard, hand, and library for any number of cards with that
// name and exile them. Then that player shuffles." The name is asked for as
// the spell resolves; the engine takes a card name as free text, so "a
// creature card name" is kept where the name is read: only a creature card by
// that name is found, and a name no creature card has finds nothing, as a name
// no legal choice could make (Day of the Moon). The searcher is the caster
// and may leave matches in the hidden zones (CR 701.23); the opponent
// shuffles whatever was taken.
func OldB321InfiniteObliteration(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		victim := TargetedPlayerOf(ctx)
		if victim == uuid.Nil {
			return nil
		}
		ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
			label+" — choose a creature card name", WithCarry(ctx.ItemTail(oldB321Named), victim))
		return nil
	}
}

// oldB321Named is the search once the name (the tail's Word) is known. An
// on-disk key: never renamed, never reused.
var oldB321Named = CardTail("card-name/s82-old-b321-infinite-obliteration", func(ctx *Context, o game.TailOutcome) error {
	name := strings.TrimSpace(o.Tail.Word)
	victim := TailCarry0(o)
	if victim == uuid.Nil {
		return nil
	}
	return SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return name != "" && !c.FaceDown && c.IsCreature() && game.CardNameMatches(c, name)
		},
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Infinite Obliteration — search their graveyard, hand and library for creature cards named " + name,
	}.Apply(ctx)
})

var (
	// "Whenever a creature dealt damage by that creature dies this turn": the
	// creature named by CondParams.Object (Touch of Moonglove's target, with its
	// CR 400.7 epoch) dealt damage to the creature that died this turn, which
	// died as a creature (CR 603.10a). Each instance carries the dead creature
	// as its payload. An on-disk key: never renamed, never reused.
	oldB321DealtDamageByDiesCondition = game.DelayedEachCondition("dies/s82-old-b321-dealt-damage-by", oldB321DealtDamageByDied)

	// "Its controller loses 2 life", read from the dead creature's last-known
	// record (CR 608.2h). An on-disk key: never renamed, never reused.
	oldB321ItsControllerLosesTwoBody = game.SimpleDelayedBody("s82-old-b321/its-controller-loses-2", oldB321ItsControllerLosesTwo)
)

func oldB321DealtDamageByDied(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if p.Object.ID == uuid.Nil {
		return nil
	}
	if _, ok := DiedCreature(ev, g); !ok {
		return nil
	}
	if !DealtDamageByThisTurn(g, p.Object.ID, ev.CardID) {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

func oldB321ItsControllerLosesTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := PayloadObject(ctx)
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return g.ChangePlayerLifeForEffect(ctx.Source(), info.Controller, -2)
}

// OldB321TouchOfMoonglove is "Target creature you control gets +1/+0 and gains
// deathtouch until end of turn. Whenever a creature dealt damage by that
// creature dies this turn, its controller loses 2 life." The pump and the
// keyword are fixed as the spell resolves (CR 611.2c); the second sentence is
// a delayed triggered ability the spell creates (CR 603.7a), controlled by the
// caster, about the target as the object it is now, firing for every creature
// that dies this turn after taking its damage (CR 603.7b).
func OldB321TouchOfMoonglove(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		creature, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := BattlefieldCardForEffect(ctx.Game, creature)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: creature, Power: 1, Label: label + " — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: creature, Keywords: []string{"deathtouch"}, Label: label + " — deathtouch"}).Apply(ctx); err != nil {
			return err
		}
		return DelayedOnEvent{
			Label:      label + " — its controller loses 2 life",
			On:         []game.EventKind{game.EventLTB},
			Condition:  oldB321DealtDamageByDiesCondition,
			CondParams: game.EffectParams{Object: game.ObjectRef{ID: creature, Epoch: c.ObjectEpoch}},
			Body:       oldB321ItsControllerLosesTwoBody,
			Repeats:    true,
		}.Apply(ctx)
	}
}
