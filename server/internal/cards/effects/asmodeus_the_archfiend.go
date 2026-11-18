package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Asmodeus the Archfiend — {4}{B}{B} Legendary Creature — Devil God 6/6
// (Adventures in the Forgotten Realms):
//
//	"Binding Contract — If you would draw a card, exile the top card of your
//	 library face down instead.
//	 {B}{B}{B}: Draw seven cards.
//	 {B}: Return all cards exiled with Asmodeus to their owner's hand and you
//	 lose that much life."
//
// The first ability is a draw replacement (CR 614.11a, 121.2) in Underrealm Lich's
// shape (DrawInstead): it cancels the draw, so nothing is "drawn" (no draw
// trigger fires), and exiles the top card of the controller's library face down
// (CR 406.3a: nobody may look at it) linked to THIS Asmodeus as the object it is
// now (CR 607.2a, 400.7), through ExileFaceDownLinkedForEffect. It replaces each
// of a "draw seven" separately (CR 121.2), and a draw step's draw; an empty
// library exiles nothing and, as nothing was drawn, loses no one the game. The
// second ability is the ordinary draw, so it exiles seven cards. The third reads
// the cards linked to the source object: each goes to ITS OWNER's hand (CR
// 400.3) as one event, and the life lost is the number of cards that really
// reached a hand; it is read from the object the ability came from, so an
// Asmodeus that left in response still returns what it exiled, and one that left
// and came back is a new object that exiled nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e4032f6-9741-4845-9325-781c9e489172",
		Name:         "Asmodeus the Archfiend",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("asmodeus-the-archfiend", asmodeusExileInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Asmodeus the Archfiend — Binding Contract: if you would draw a card, exile the top card of your library face down instead",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{B}{B}{B}: Draw seven cards.",
				Purpose: game.Purpose{Draws: 7, Answers: game.AnswerValue},
				Cost:    ManaCost("{B}{B}{B}"),
				Effect:  Do(DrawCards{N: 7}),
			},
			{
				Label:   "{B}: Return all cards exiled with Asmodeus to their owner's hand and you lose that much life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{B}"),
				Effect:  asmodeusReturn,
			},
		},
	})
}

// asmodeusExileInstead is the replacement's body: the top card goes to exile
// face down, linked to Asmodeus, and the rest of the draw instruction goes on.
func asmodeusExileInstead(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	if p := g.PlayerByIDForEffect(drawer); p != nil && p.Library != nil && len(p.Library.Cards) > 0 {
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		if _, err := g.ExileFaceDownLinkedForEffect(source, drawer, top); err != nil {
			return err
		}
	}
	return done.Run(g)
}

// asmodeusReturn returns every card exiled with this Asmodeus to its owner's
// hand; the life lost is told by the tail from what reached a hand.
func asmodeusReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
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
	return g.BounceCardsToHandKeyedForEffect(ids, ctx.Tail(asmodeusReturned))
}

// asmodeusReturned is "and you lose that much life": one life per card that
// reached a hand. An on-disk key: never renamed, never reused.
var asmodeusReturned = CardTail("asmodeus-the-archfiend/returned", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -len(o.IDs))
})
