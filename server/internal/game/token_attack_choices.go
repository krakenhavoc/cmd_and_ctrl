package game

import (
	"slices"
	"strconv"

	"github.com/google/uuid"
)

// chooseTokenAttackGroupsLocked allocates the settled count, after token
// doublers, among the named opponent and their planeswalkers. It is not
// targeting, and battles are not permitted by myriad (CR 702.116a).
func (g *Game) chooseTokenAttackGroupsLocked(ev *ReplacementEvent, index int, chosen []TokenGroup) error {
	if index == len(ev.TokenGroups) {
		out := *ev
		out.TokenGroups = chosen
		return g.applyResolvedTokenCreationLocked(&out)
	}
	grp := ev.TokenGroups[index]
	if grp.AttackOpponent == uuid.Nil {
		return g.chooseTokenAttackGroupsLocked(ev, index+1, append(slices.Clone(chosen), grp))
	}
	var targets []uuid.UUID
	for _, ref := range g.AttackTargetsForEffect(ev.TokenController) {
		if ref.Kind != AttackTargetBattle && g.defendingPlayerForAttackLocked(ref.ID) == grp.AttackOpponent {
			targets = append(targets, ref.ID)
		}
	}
	return g.chooseTokenAttackSharesLocked(ev, index, chosen, targets, grp.Count)
}

func (g *Game) chooseTokenAttackSharesLocked(ev *ReplacementEvent, index int, chosen []TokenGroup, targets []uuid.UUID, left int) error {
	if left <= 0 {
		return g.chooseTokenAttackGroupsLocked(ev, index+1, chosen)
	}
	grp := ev.TokenGroups[index]
	part := func(n int) []TokenGroup {
		out := slices.Clone(chosen)
		if n > 0 {
			share := grp
			share.Count, share.AttackOpponent = n, uuid.Nil
			share.Attacking = uuid.Nil
			if len(targets) > 0 {
				share.Attacking = targets[0]
			}
			out = append(out, share)
		}
		return out
	}
	if len(targets) <= 1 {
		return g.chooseTokenAttackGroupsLocked(ev, index+1, part(left))
	}
	name := "player"
	if p := g.playerByIDLocked(targets[0]); p != nil {
		name = p.Name
	} else if c := findBattlefieldCard(g, targets[0]); c != nil {
		name = c.Name
	}
	return g.QueueChooseNumberForEffect(ChooseNumber{
		Chooser: ev.TokenController, Source: ev.Source, Resource: PayResourceNone,
		Max: left, Goal: left,
		Question: ev.TokenAttackChoicesLabel + " — how many of the " + strconv.Itoa(left) + " copies attack " + name + "?",
		Then: func(g *Game, n int) error {
			return g.chooseTokenAttackSharesLocked(ev, index, part(n), targets[1:], left-n)
		},
	})
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
