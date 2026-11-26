package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB501CycledOrDies is one triggered ability with two conditions: "When
// you cycle this card and when this creature dies" (the Alara Reborn
// Sojourners). It watches from the graveyard, where the cycled card already is
// (CR 702.29c) and where the creature has just arrived (CR 700.4), and so
// fires once for either event.
func S82OldB501CycledOrDies(label string, effect Effect) game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventCycle, game.EventLTB},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventCycle {
				return Self(ev, source, lki, g)
			}
			return ThisDied(ev, source, lki, g)
		}, label, effect)
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

// S82OldB501GlassdustHulk is "this creature gets +1/+1 until end of turn and
// can't be blocked this turn": both halves land on the source only while it is
// still the permanent the ability came from (CR 400.7).
func S82OldB501GlassdustHulk(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := NewContext(g, item)
		if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
	}
}

// S82OldB501SovereignsSearch is Sovereigns of Lost Alara's search: the lone
// attacker (read off the triggering event) is the creature the Aura must be
// able to enchant, and the Aura enters attached to it. Nothing happens when the
// attacker has left the battlefield.
func S82OldB501SovereignsSearch(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !OnBattlefield(g, attacker) {
		return nil
	}
	return SearchLibrary{
		Player:           item.Controller,
		Predicate:        S82OldB431AuraThatCouldEnchant(g, item.Controller, attacker),
		Dest:             game.ZoneBattlefield,
		Limit:            1,
		Shuffle:          true,
		Optional:         true,
		Reason:           "Sovereigns of Lost Alara — an Aura card that could enchant that creature",
		EntersAttachedTo: game.TargetRef{Kind: game.TargetCard, ID: attacker},
	}.Apply(NewContext(g, item))
}

// S82OldB501YouSacrificedACreature is "whenever you sacrifice a creature". The
// sacrifice event fires before the zone move, so the permanent is still on the
// battlefield to be read.
func S82OldB501YouSacrificedACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

const s82OldB501SagesLabel = "Sages of the Anima — reveal the top three cards, put the creature cards into your hand and the rest on the bottom"

// S82OldB501SagesInstead is the replaced draw of Sages of the Anima: reveal the
// top three cards, every creature card among them goes to the hand, and the
// rest go on the bottom in any order the player chooses. The rest of the draw
// instruction waits for all of it (the parked DrawDone).
func S82OldB501SagesInstead(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	looked := g.RevealTopOfLibraryForEffect(drawer, source, 3, s82OldB501SagesLabel)
	var picked []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() && !c.IsToken() {
			picked = append(picked, id)
		}
	}
	tail := game.KeyedTail{Ref: s82OldB501SagesTaken, Player: drawer, Source: source,
		Carry: append([]uuid.UUID{done.Park(g)}, looked...)}
	return FinishTake(g, source, drawer, false, s82OldB501SagesLabel, tail, picked)
}

// s82OldB501SagesTaken puts the rest of the reveal on the bottom in the
// player's order, then pays the parked draw. An on-disk key.
var s82OldB501SagesTaken = game.RegisterTail("sages-of-the-anima/taken", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	res := TakeResult(g, o.Tail.Source, o.Tail.Player, o.Tail.Carry[1:], o.IDs)
	return TakeThenErr(g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    LibraryOrderLabel(g, res.Source, "put the rest on the bottom of your library in any order"),
		ThenTail:  game.KeyedTail{Ref: s82OldB501SagesDone, Player: res.Player, Source: res.Source, Carry: o.Tail.Carry[:1]},
	}))
})

// s82OldB501SagesDone pays the rest of the replaced draw once the order is
// settled. An on-disk key.
var s82OldB501SagesDone = game.RegisterTail("sages-of-the-anima/done", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(o.Tail.Carry[0]).Run(g)
})
