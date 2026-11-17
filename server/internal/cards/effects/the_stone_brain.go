package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b28dcb41-5e47-4719-892f-da009a5d50b4",
		Name:         "The Stone Brain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}, Exile The Stone Brain: Choose a card name. Search target opponent's graveyard, hand, and library for up to four cards with that name and exile them. That player shuffles, then draws a card for each card exiled from their hand this way. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost(), ExileThis()),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				victim := firstLegalTargetPlayer(ctx)
				if victim == uuid.Nil {
					return nil
				}
				g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
					"The Stone Brain — choose a card name",
					withCarry(ctx.ItemTail(theStoneBrainNamed), victim))
				return nil
			},
		}},
	})
}

// theStoneBrainNamed is the search once the name (the tail's Word) is
// known. An on-disk key.
var theStoneBrainNamed = CardTail("then/the-stone-brain-named", func(ctx *Context, o game.TailOutcome) error {
	victim, name := tailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return SearchLibrary{
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
		ThenTail:             withCarry(ctx.ItemTail(theStoneBrainDraw), victim),
	}.Apply(ctx)
})

// theStoneBrainDraw is "that player draws a card for each card exiled from
// their hand this way", after the shuffle. An on-disk key.
var theStoneBrainDraw = CardTail("then/the-stone-brain-draw", func(ctx *Context, o game.TailOutcome) error {
	victim := tailCarry0(o)
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
	return DrawCards{Player: victim, N: n}.Apply(ctx)
})
