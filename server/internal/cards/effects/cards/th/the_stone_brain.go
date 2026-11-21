package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Stone Brain — Legendary Artifact {2} (The Brothers' War):
//
//	"{2}, {T}, Exile The Stone Brain: Choose a card name. Search target
//	 opponent's graveyard, hand, and library for up to four cards with that
//	 name and exile them. That player shuffles, then draws a card for each
//	 card exiled from their hand this way. Activate only as a sorcery."
//
// Ancient Vendetta's two steps (the name asked as the ability resolves,
// then the extraction family's three-zone search capped at four), with the
// draw after the shuffle: the opponent draws one card for each exiled card
// that was in their hand, told apart by the zone move that took it.
// A target opponent who is no longer legal makes the ability do nothing
// (CR 608.2b); the exile is a cost and stands regardless.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b28dcb41-5e47-4719-892f-da009a5d50b4",
		Name:         "The Stone Brain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Exile The Stone Brain: Choose a card name. Search target opponent's graveyard, hand, and library for up to four cards with that name and exile them. That player shuffles, then draws a card for each card exiled from their hand this way. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim := effects.FirstLegalTargetPlayer(ctx)
				if victim == uuid.Nil {
					return nil
				}
				g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
					"The Stone Brain — choose a card name",
					effects.WithCarry(ctx.ItemTail(theStoneBrainNamed), victim))
				return nil
			},
		}},
	})
}

// theStoneBrainNamed is the search once the name (the tail's Word) is
// known. An on-disk key.
var theStoneBrainNamed = effects.CardTail("then/the-stone-brain-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim, name := effects.TailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return effects.SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return game.CardNameMatches(c, name)
		},
		Dest:                 game.ZoneExile,
		Limit:                4,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "The Stone Brain — search their graveyard, hand and library for up to four cards named " + name,
		ThenTail:             effects.WithCarry(ctx.ItemTail(theStoneBrainDraw), victim),
	}.Apply(ctx)
})

// theStoneBrainDraw is "that player draws a card for each card exiled from
// their hand this way", after the shuffle. An on-disk key.
var theStoneBrainDraw = effects.CardTail("then/the-stone-brain-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	if victim == uuid.Nil || len(o.IDs) == 0 {
		return nil
	}
	fromHand := map[uuid.UUID]bool{}
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventZoneMove && ev.OldZone == game.ZoneHand && ev.NewZone == game.ZoneExile {
			fromHand[ev.CardID] = true
		}
	}
	n := 0
	for _, id := range o.IDs {
		if fromHand[id] {
			n++
		}
	}
	return effects.DrawCards{Player: victim, N: n}.Apply(ctx)
})
