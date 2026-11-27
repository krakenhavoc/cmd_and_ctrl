package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b627_helpers.go — Judgment cards, S82 OLD batch 627 (#5611).

// S82OldB627CounterExileAndMayPlay is Spelljack's body: counter the spell the
// ability targets, exile it instead of putting it into its owner's graveyard,
// and give the caster a standing {0} permission over it for as long as it
// remains exiled (Kheru Spellsnatcher's shape). Only a spell that really
// reached exile was "countered this way" (a spell that can't be countered
// stays on the stack and gives nothing).
func S82OldB627CounterExileAndMayPlay(name string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		g := ctx.Game
		legal := ctx.LegalTargets()
		if len(legal) == 0 || legal[0].Kind != game.TargetCard {
			return nil
		}
		stackID := legal[0].ID
		if g.StackItemForEffect(stackID) == nil {
			return nil
		}
		if err := (CounterTarget{StackID: stackID, Dest: game.ZoneRef{Kind: game.ZoneExile}}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(stackID); z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		c, ok := g.LookupCardForEffect(stackID)
		if !ok {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     item.Controller,
			Zone:       game.ZoneExile,
			Cost:       "{0}",
			CastOnly:   true,
			Duration:   game.WhileInZoneDuration(),
			Source:     item.SourceCardID,
			SourceName: name,
			Label:      name + " — you may play it without paying its mana cost for as long as it remains exiled",
		}, []game.Card{c})
		return nil
	}
}

// S82OldB627ExchangeGraveyardAndLibrary is "Exchange your graveyard and
// library. Then shuffle your library." (Morality Shift): the library cards go
// to the graveyard as one move, then the cards that were in the graveyard
// before it go to the library as another, and the library is shuffled. A
// replacement that pauses either move carries the rest as data (the middle
// step is a registered tail).
func S82OldB627ExchangeGraveyardAndLibrary(ctx *Context) error {
	g := ctx.Game
	p := g.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Library == nil || p.Graveyard == nil {
		return nil
	}
	lib := make([]uuid.UUID, 0, p.Library.Size())
	for _, c := range p.Library.Cards {
		lib = append(lib, c.InstanceID)
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(lib, ctx.Tail(moralityShiftLibraryMoved))
}

// moralityShiftLibraryMoved is the second step: every graveyard card that did
// not just arrive from the library goes to the library, then the shuffle. An
// on-disk key: never renamed, never reused.
var moralityShiftLibraryMoved = CardTail("tail/s82-old-b627-morality-shift-library-moved", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	p := g.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	arrived := make(map[uuid.UUID]bool, len(o.IDs))
	for _, id := range o.IDs {
		arrived[id] = true
	}
	var before []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if !arrived[c.InstanceID] {
			before = append(before, c.InstanceID)
		}
	}
	return g.TuckCardsToLibraryKeyedForEffect(before, game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: ctx.Controller()})
})

// S82OldB627ReturnLinkedExiledCardsToHand is the leaves-the-battlefield half of
// Wormfang Behemoth: "When this creature leaves the battlefield, return the
// exiled cards to their owner's hand." The two abilities are linked (CR
// 607.2a): "the exiled cards" are whatever the enters trigger exiled with THIS
// object, read off the leaving object's link as it last existed (CR 603.10). A
// Behemoth that left before its enters trigger resolved has already had this
// trigger resolve with nothing to return, so what it exiles afterwards stays
// exiled. A card that has since left exile is not pulled back (CR 400.7).
func S82OldB627ReturnLinkedExiledCardsToHand(label string) game.TriggeredAbility {
	return WhenThisLeaves(label, func(g *game.Game, item *game.StackItem) error {
		ref, ok := NewContext(g, item).SourceRef()
		if !ok {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			ids = append(ids, c.InstanceID)
		}
		if len(ids) == 0 {
			return nil
		}
		return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
	})
}
