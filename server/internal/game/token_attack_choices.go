package game

import (
	"fmt"

	"github.com/google/uuid"
)

// startSettledSplitLocked sends the groups of a settled creation to what
// they attack: each group that names an opponent (TokenGroup.AttackOpponent)
// has its settled count, after token doublers, allocated among that
// opponent and their planeswalkers. It is not targeting, and battles are not
// permitted by myriad (CR 702.116a).
//
// The questions are the engine's attack split (attack_split.go), with the
// settled creation as its data: the groups, the one being split, the ones
// placed, and the creation's tail as a registered key. The window has
// settled by now, so the split is not part of the CR 614 pipeline and asks
// no replacement anything; a table waiting on a question is a restore
// point (#4294). A creation that is split carries a keyed tail, or none:
// a closure is not data.
//
// Caller must hold g.mu.
func (g *Game) startSettledSplitLocked(ev *ReplacementEvent) error {
	var tail KeyedTail
	if t := ev.tokenTail; t != nil {
		if t.then != nil || t.report != nil {
			return fmt.Errorf("%w: a creation split between opponents needs a keyed tail", ErrInvalidParam)
		}
		tail = copyKeyedTail(t.keyed)
	}
	return g.continueSettledSplitLocked(AttackSplit{
		Controller: ev.TokenController,
		Source:     ev.Source,
		Label:      ev.TokenAttackChoicesLabel,
		Tail:       tail,
		settled: &settledCreation{
			Groups:    cloneTokenGroups(ev.TokenGroups),
			Attacking: ev.TokenAttacking,
			Together:  ev.TokenTogether,
		},
	})
}

// continueSettledSplitLocked moves on to the next group that names an
// opponent, or makes the tokens once there is none.
//
// Caller must hold g.mu.
func (g *Game) continueSettledSplitLocked(s AttackSplit) error {
	st := s.settled
	for st.Index < len(st.Groups) {
		grp := st.Groups[st.Index]
		if grp.AttackOpponent == uuid.Nil {
			st.Chosen = append(st.Chosen, grp)
			st.Index++
			continue
		}
		var targets []uuid.UUID
		for _, ref := range g.AttackTargetsForEffect(s.Controller) {
			if ref.Kind != AttackTargetBattle && g.defendingPlayerForAttackLocked(ref.ID) == grp.AttackOpponent {
				targets = append(targets, ref.ID)
			}
		}
		s.Targets, s.Shares, s.Count = targets, nil, grp.Count
		return g.advanceAttackSplitLocked(s)
	}
	out := &ReplacementEvent{
		Kind:                    RepEventCreateTokens,
		Actor:                   s.Controller,
		Source:                  s.Source,
		TokenController:         s.Controller,
		TokenGroups:             st.Chosen,
		TokenAttacking:          st.Attacking,
		TokenTogether:           st.Together,
		TokenAttackChoicesLabel: s.Label,
	}
	if !s.Tail.IsZero() {
		out.tokenTail = &tokenTail{keyed: s.Tail}
	}
	return g.applyResolvedTokenCreationLocked(out)
}

// placeSettledGroupLocked records the shares of the group being split as
// groups of their own, each attacking the thing it was sent to (or none
// attacking when there is nothing to attack, CR 508.4a), and goes on.
//
// Caller must hold g.mu.
func (g *Game) placeSettledGroupLocked(s AttackSplit) error {
	st := s.settled
	grp := st.Groups[st.Index]
	for i, n := range s.Shares {
		if n <= 0 {
			continue
		}
		share := grp
		share.Count, share.AttackOpponent, share.Attacking = n, uuid.Nil, uuid.Nil
		if i < len(s.Targets) {
			share.Attacking = s.Targets[i]
		}
		st.Chosen = append(st.Chosen, share)
	}
	st.Index++
	s.Targets, s.Shares, s.Count = nil, nil, 0
	return g.continueSettledSplitLocked(s)
}

func (g *Game) enterCreatedTokensTogetherLocked(tokens []stagedToken, source uuid.UUID, done func(*Game, []uuid.UUID) error) error {
	b := &entryBatch{source: source}
	ids := make([]uuid.UUID, 0, len(tokens))
	for _, tok := range tokens {
		ids = append(ids, tok.card.InstanceID)
		g.enteringTokens = append(g.enteringTokens, tok.card)
		b.members = append(b.members, entryBatchMember{cardID: tok.card.InstanceID, controller: tok.card.Controller,
			token: true, tokenEntry: tok.entry, attacking: tok.card.AttackingTarget})
	}
	b.then = func(g *Game, entered []uuid.UUID) error {
		for _, id := range ids {
			g.dropEnteringTokenLocked(id)
		}
		if done != nil {
			return done(g, entered)
		}
		return nil
	}
	return g.continueEntryBatchLocked(b)
}
