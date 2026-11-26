package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b337_helpers.go — Fate Reforged, batch 337 (#5552): the manifest
// cards that put counters on the manifested card, and Write into Being.

// S82OldB337ManifestThenCounters is "Manifest the top card of your library,
// then put <N | X> +1/+1 counters on it" (Fierce Invocation, Wildcall,
// Formless Nurturing; Wildcall passes its announced X). The card enters through the keyed face-down entry, so
// the counters go on from its tail once the permanent has arrived, even when a
// replacement window asked a question first. An empty library manifests
// nothing and puts no counters (CR 701.40a).
type S82OldB337ManifestThenCounters struct {
	N int
}

func (m S82OldB337ManifestThenCounters) Apply(ctx *Context) error {
	n := m.N
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{top},
		game.ZoneEntryOptions{Controller: ctx.Controller(), FaceDown: game.FaceDownManifested},
		WithN(ctx.ItemTail(s82OldB337Counters), n))
}

// s82OldB337Counters puts the carried +1/+1 counters on what was manifested.
// An on-disk key: never renamed, never reused.
var s82OldB337Counters = CardTail("then/s82-old-b337-manifest-counters", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneBattlefield {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: o.Tail.N}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})

// S82OldB337WriteIntoBeing is "Look at the top two cards of your library.
// Manifest one of those cards, then put the other on the top or bottom of your
// library." Only the looker sees the pair (CR 701.20a, a look), picks which to
// manifest, and the manifested card enters before the other is placed. With
// one card in the library it is manifested and nothing is placed.
type S82OldB337WriteIntoBeing struct{}

func (S82OldB337WriteIntoBeing) Apply(ctx *Context) error {
	player := ctx.Controller()
	looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 2)
	switch len(looked) {
	case 0:
		return nil
	case 1:
		return s82OldB337Manifest(ctx.Game, player, ctx.Source(), looked[0], uuid.Nil)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Write into Being — manifest one of those cards; the other goes on the top or bottom of your library",
		Cards:    looked,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneLibrary,
		ThenKey:  s82OldB337WritePick,
		Carry:    looked,
	})
	return nil
}

// s82OldB337WritePick is the pick's answer. An on-disk key: never renamed,
// never reused.
var s82OldB337WritePick = game.RegisterChooseCardsThen("choose-cards/s82-old-b337-write-into-being", func(g *game.Game, r game.CardsChosen) error {
	chosen := uuid.Nil
	if r.Answered && len(r.Picked) > 0 {
		chosen = r.Picked[0]
	} else if len(r.Carry) > 0 {
		chosen = r.Carry[0]
	}
	other := uuid.Nil
	for _, id := range r.Carry {
		if id != chosen {
			other = id
		}
	}
	return s82OldB337Manifest(g, r.Chooser, r.Source, chosen, other)
})

// s82OldB337Manifest manifests `chosen` and, from the entry's tail, puts
// `other` (when it is still in the library) on the top or bottom.
func s82OldB337Manifest(g *game.Game, player, source, chosen, other uuid.UUID) error {
	if z := g.FindCardZoneForEffect(chosen); z == nil || z.Kind != game.ZoneLibrary {
		return s82OldB337Place(g, player, source, other)
	}
	tail := KeyedContext(g, player, source).Tail(s82OldB337WritePlace)
	if other != uuid.Nil {
		tail.Carry = []uuid.UUID{other}
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{chosen},
		game.ZoneEntryOptions{Controller: player, FaceDown: game.FaceDownManifested}, tail)
}

var s82OldB337WritePlace = CardTail("then/s82-old-b337-write-into-being-place", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return s82OldB337Place(ctx.Game, ctx.Controller(), ctx.Source(), o.Tail.Carry[0])
})

func s82OldB337Place(g *game.Game, player, source, card uuid.UUID) error {
	if card == uuid.Nil {
		return nil
	}
	z := g.FindCardZoneForEffect(card)
	if z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return PutInLibraryInAnyOrder{
		Chooser:   player,
		Cards:     []uuid.UUID{card},
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTopOrBottom,
		Label:     "Write into Being — put the other card on the top or bottom of your library",
	}.Apply(KeyedContext(g, player, source))
}
