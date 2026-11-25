package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unmoored Ego — Sorcery {1}{U}{B} (Guilds of Ravnica):
//
//	"Choose a card name. Search target opponent's graveyard, hand, and
//	 library for up to four cards with that name and exile them. That player
//	 shuffles, then draws a card for each card exiled from their hand this
//	 way."
//
// Ancient Vendetta's name-and-search (the caster names a card as it
// resolves, then searches the opponent's three zones for up to four), with
// Deadly Cover-Up's draw: the opponent shuffles first, then draws once for
// each card the search moved from their hand to exile, counted from the zone
// moves themselves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1206b672-e075-45ba-97d4-7c3012be062d",
		Name:         "Unmoored Ego",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Unmoored Ego — choose a card name",
				effects.WithCarry(ctx.ItemTail(unmooredEgoNamed), victim))
			return nil
		},
	})
}

// unmooredEgoNamed is the search once the name (the tail's Word) is known.
// An on-disk key.
var unmooredEgoNamed = effects.CardTail("then/unmoored-ego-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim, name := effects.TailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return effects.SearchLibrary{
		Player:               ctx.Controller(),
		LibraryOwner:         victim,
		Predicate:            func(c game.Card) bool { return game.CardNameMatches(c, name) },
		Dest:                 game.ZoneExile,
		Limit:                4,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Unmoored Ego — search their graveyard, hand and library for up to four cards named " + name,
		ThenTail:             effects.WithCarry(ctx.ItemTail(unmooredEgoDraw), victim),
	}.Apply(ctx)
})

// unmooredEgoDraw is "that player draws a card for each card exiled from
// their hand this way", counted from the zone moves the search made. An
// on-disk key.
var unmooredEgoDraw = effects.CardTail("then/unmoored-ego-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	if victim == uuid.Nil || len(o.IDs) == 0 {
		return nil
	}
	found := map[uuid.UUID]bool{}
	for _, id := range o.IDs {
		found[id] = true
	}
	n := 0
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventZoneMove && found[ev.CardID] && ev.OldZone == game.ZoneHand && ev.NewZone == game.ZoneExile {
			n++
			delete(found, ev.CardID)
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: victim, N: n}.Apply(ctx)
})
