package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Echo of Eons — Sorcery {4}{U}{U}:
//
//	"Each player shuffles their hand and graveyard into their library,
//	 then draws seven cards.
//	 Flashback {2}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Timetwister at the table, on Wheel of Fortune's two-phase shape:
// every player's hand and graveyard go into their own library FIRST —
// one simultaneous exit per player, through the same
// TuckCardsToLibraryThenForEffect batch Jace, the Mind Sculptor's
// ultimate already chains exile→tuck→shuffle through — and only once
// EVERY player's shuffle has landed does anyone draw. #539's "the
// commander replacement holds 'from anywhere'" generalization is what makes the graveyard
// half of the tuck real: the same exit primitive that moves a hand
// card to a library moves a graveyard card there too, so "hand and
// graveyard" is one combined batch per player rather than two
// mechanisms bolted together. A commander sitting in a hand or
// graveyard gets its CR 903.9b offer exactly as any other tuck does:
// the library is one of that replacement's destinations, so the
// owner is asked before the card moves.
//
// The shuffle-then-draw split (not "shuffle a player, draw for that
// player, shuffle the next") mirrors Wheel of Fortune's own reasoning
// one step over: every player's cards leave together before anyone's
// new hand is dealt, so nothing downstream can read a player's
// post-Echo hand before the whole table has re-shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "23d3e5fe-3f82-44cf-91a1-6646a12a0255",
		Name:             "Echo of Eons",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			return echoOfEonsShuffleEach(ctx.Game, players, players)
		},
	})
}

// echoOfEonsShuffleEach tucks the head player's hand and graveyard
// into their own library as one simultaneous exit, shuffles, and
// continues with the tail — recursive rather than a plain loop
// because the tuck can pause on a CR 903.9b prompt, and the
// continuation must resolve against whatever *Game it is handed then,
// not the one this call started with. Once every player has shuffled,
// `all` (the untouched full roster) hands off to the draw phase.
//
// Package-level and capture-free: nothing but player IDs crosses a
// closure boundary, the same discipline jaceUltimate
// (jace_the_mind_sculptor.go) follows for the identical reason — undo
// resolves a delayed continuation against a cloned game.
func echoOfEonsShuffleEach(g *game.Game, remaining, all []uuid.UUID) error {
	if len(remaining) == 0 {
		return echoOfEonsDrawEach(g, all)
	}
	id, rest := remaining[0], remaining[1:]
	ids := append(allHandCardIDs(g, id), echoOfEonsGraveyardCardIDs(g, id)...)
	return g.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
		if err := g.ShuffleLibraryForEffect(id); err != nil {
			return err
		}
		return echoOfEonsShuffleEach(g, rest, all)
	})
}

// echoOfEonsDrawEach is the second phase: every player draws seven,
// once every shuffle above has landed. A draw does not pause on
// anything a commander would care about, so a plain loop is enough
// here where the shuffle phase needed recursion.
func echoOfEonsDrawEach(g *game.Game, players []uuid.UUID) error {
	for _, id := range players {
		if err := g.DrawNForEffect(id, 7); err != nil {
			return err
		}
	}
	return nil
}

// echoOfEonsGraveyardCardIDs is every card in a player's graveyard, in
// zone order — the graveyard half of "shuffles their hand and
// graveyard into their library", alongside allHandCardIDs's hand half.
func echoOfEonsGraveyardCardIDs(g *game.Game, player uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}
