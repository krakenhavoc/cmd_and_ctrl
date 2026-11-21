package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Indomitable Creativity — Sorcery {X}{R}{R}{R} (Breaking News, #5929):
//
//	"Destroy X target artifacts and/or creatures. For each permanent
//	 destroyed this way, its controller reveals cards from the top of their
//	 library until an artifact or creature card is revealed and exiles that
//	 card. Those players put the exiled cards onto the battlefield, then
//	 shuffle."
//
// Exactly X targets (CountFromX), each still legal as it resolves (CR
// 608.2b), destroyed together as one simultaneous event; an indestructible
// permanent is not destroyed and so gets no reveal. The rest is the
// destroy's keyed tail, told the permanents that really died: for each, in
// the order they were chosen, its controller as it last existed (CR 608.2h)
// reveals from the top of their library until an artifact or creature card
// (a token is not a card, CR 111.8) that no earlier reveal of that player
// has already taken, and that card is exiled. A controller of two destroyed
// permanents reveals twice, the second reveal starting from what is left. A
// library with no such card is revealed entirely and exiles nothing. Then
// every exiled card is put onto the battlefield as one simultaneous entry,
// under its owner's control (each is its revealer's own card), and then each
// player who revealed shuffles, in turn order from the active player (CR
// 101.4); the cards revealed and not exiled never left the library, so the
// shuffle is what orders them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2161ae2-e5d2-4a56-85d4-d2214c5b3e2e",
		Name:         "Indomitable Creativity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetsCountedByX(effects.TargetPermanent("X target artifacts and/or creatures", effects.Or(effects.Artifact(), effects.Creature()))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			ids = ctx.Game.DestructibleForEffect(ids)
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect(ids, ctx.Tail(indomitableCreativityReveal))
		},
	})
}

// indomitableCreativityReveal is the reveal, the exile and the entry, keyed
// (#4253). An on-disk key: never renamed, never reused.
var indomitableCreativityReveal = effects.MassTail("mass/indomitable-creativity-reveal", func(ctx *effects.Context, swept []game.Card, _ int) error {
	g := ctx.Game
	taken := map[uuid.UUID]bool{}
	var exiled []uuid.UUID
	revealers := map[uuid.UUID]bool{}
	for _, dead := range swept {
		player := dead.Controller
		p := g.PlayerByIDForEffect(player)
		if p == nil || p.Eliminated || p.Library == nil {
			continue
		}
		revealers[player] = true
		var run []uuid.UUID
		hit := uuid.Nil
		for i := len(p.Library.Cards) - 1; i >= 0; i-- {
			c := p.Library.Cards[i]
			if taken[c.InstanceID] {
				continue
			}
			run = append(run, c.InstanceID)
			if !c.IsToken() && (c.IsArtifact() || c.IsCreature()) {
				hit = c.InstanceID
				break
			}
		}
		if len(run) > 0 {
			g.RevealForEffect(game.RevealSpec{
				Player: player,
				Source: ctx.Source(),
				Reason: "Indomitable Creativity — reveal until an artifact or creature card",
				Cards:  run,
			})
		}
		if hit != uuid.Nil {
			taken[hit] = true
			exiled = append(exiled, hit)
		}
	}
	var players []uuid.UUID
	for _, seat := range effects.ApnapLivingSeats(g) {
		if revealers[seat] {
			players = append(players, seat)
		}
	}
	tail := effects.WithCarry(ctx.Tail(indomitableCreativityShuffle), players...)
	if len(exiled) == 0 {
		return tail.Run(g, nil)
	}
	g.ExileCardsForEffect(exiled)
	entries := make([]game.BatchEntry, 0, len(exiled))
	for _, id := range exiled {
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{}, tail)
})

// indomitableCreativityShuffle is "then shuffle": every player who revealed.
// An on-disk key: never renamed, never reused.
var indomitableCreativityShuffle = effects.CardTail("then/indomitable-creativity-shuffle", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, seat := range o.Tail.Carry {
		if err := ctx.Game.ShuffleLibraryForEffect(seat); err != nil {
			return err
		}
	}
	return nil
})
