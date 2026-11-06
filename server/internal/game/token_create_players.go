package game

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// token_create_players.go — "each player creates a token" (#4867, ADR 0061
// amendment 2026-11-06): one effect in which several players create tokens
// at the same time. Fractured Identity's "each player other than its
// controller creates a token that's a copy of it".
//
// CR 101.4 and CR 608.2f: an action several players take at the same time
// is processed simultaneously, so every player's tokens enter in ONE event
// and each newcomer sees the others enter (CR 603.6a). But each player's
// creation is still that player's own, and the replacement effects that
// apply to a token being created (CR 701.7b) apply to it alone:
// a Doubling Season doubles only the tokens its controller would create,
// and each player orders their own replacements (CR 616.1). So the door
// opens one creation window per player, in APNAP order, against a board
// where none of the tokens exists yet, mints what each window settled on,
// and enters every token through the one simultaneous entry
// (enterCreatedTokensTogetherLocked) a single creation of several tokens
// already takes.
//
// A window that asks a question (a CR 616 ordering of two of that player's
// token replacements) is not waited for: that player's creation is paused
// on the ordinary creation frame, and its tokens enter on their own once
// the question is answered, as every player's did before this door. The
// others' tokens enter together without them. That is the conservative
// fallback: it writes no new frame, and it is never worse than the
// one-creation-per-player loop it replaces.

// CreateTokensForPlayersTogetherForEffect creates, for every spec, the
// tokens it names under its own Controller, and enters every token of the
// specs whose windows settled as one simultaneous entry. A spec's
// Attacking, its groups' Attacking, AttackOpponent and Blocking are not
// supported (no "each player creates" card puts tokens onto the
// battlefield attacking or blocking) and are refused. There is no
// continuation: a caller with more to do with the tokens takes the
// single-creation doors.
//
// Caller must hold g.mu.
func (g *Game) CreateTokensForPlayersTogetherForEffect(specs []TokenCreation) error {
	for _, spec := range specs {
		if spec.Attacking != uuid.Nil {
			return fmt.Errorf("%w: a creation by several players cannot be created attacking", ErrInvalidParam)
		}
		for _, grp := range spec.Groups {
			if grp.Attacking != uuid.Nil || grp.AttackOpponent != uuid.Nil || grp.Blocking != uuid.Nil {
				return fmt.Errorf("%w: a creation by several players cannot be created attacking or blocking", ErrInvalidParam)
			}
		}
	}
	var batch []stagedToken
	source := uuid.Nil
	for _, spec := range specs {
		if g.playerByIDLocked(spec.Controller) == nil {
			continue
		}
		groups := make([]TokenGroup, 0, len(spec.Groups))
		for _, grp := range spec.Groups {
			if grp.Count > 0 {
				groups = append(groups, grp)
			}
		}
		if len(groups) == 0 {
			continue
		}
		ev := &ReplacementEvent{
			Kind:            RepEventCreateTokens,
			Actor:           spec.Controller,
			Source:          spec.Source,
			TokenController: spec.Controller,
			TokenGroups:     groups,
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			// This player's creation waits on its own question and
			// enters on its own when it is answered (createTokensLocked's
			// resume). Nothing of it exists yet.
			continue
		}
		g.clearReplacementEventLocked(ev.ID)
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			return err
		}
		if out == nil || out.Canceled {
			continue
		}
		if source == uuid.Nil {
			source = out.Source
		}
		for _, grp := range out.TokenGroups {
			for i := 0; i < grp.Count; i++ {
				batch = append(batch, stagedToken{card: g.mintTokenLocked(grp, out.TokenController), entry: grp.Entry})
			}
		}
	}
	switch len(batch) {
	case 0:
		return nil
	case 1:
		// A one-member batch is the same entry; this path keeps its frame,
		// as a single creation of one token does.
		return g.enterCreatedTokensKeyedLocked(batch, nil, source, tokenDone{})
	}
	return g.enterCreatedTokensTogetherLocked(batch, source, tokenDone{})
}
