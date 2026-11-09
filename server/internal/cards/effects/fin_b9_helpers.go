package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b9_helpers.go — Final Fantasy batches 9 and 10 (#4462): the party
// (CR 700.8), which only The Destined cards read.

// finB9PartyTypes are the four creature types of a party (CR 700.8).
var finB9PartyTypes = [4]string{"Cleric", "Rogue", "Warrior", "Wizard"}

// finB9PartySize is the number of creatures in `player`'s party (CR 700.8,
// 700.8a): up to one creature they control per party type, and a creature
// with several of the types counts for only one of them, assigned so the
// result is the highest (CR 700.8b) — a maximum bipartite matching between
// their creatures and the four types. A changeling is every type
// (HasSubtype reads CR 702.73a), and the types are read through the layers.
func finB9PartySize(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	var fits [][4]bool
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		var f [4]bool
		any := false
		for i, t := range finB9PartyTypes {
			if c.HasSubtype(t) {
				f[i], any = true, true
			}
		}
		if any {
			fits = append(fits, f)
		}
	}
	owner := [4]int{-1, -1, -1, -1} // type slot → index into fits
	var try func(c int, seen *[4]bool) bool
	try = func(c int, seen *[4]bool) bool {
		for t := 0; t < 4; t++ {
			if !fits[c][t] || seen[t] {
				continue
			}
			seen[t] = true
			if owner[t] < 0 || try(owner[t], seen) {
				owner[t] = c
				return true
			}
		}
		return false
	}
	n := 0
	for c := range fits {
		var seen [4]bool
		if try(c, &seen) {
			n++
		}
	}
	return n
}

// finB9HasFullParty is "you have a full party": four creatures in the
// party (CR 700.8c).
func finB9HasFullParty(g *game.Game, player uuid.UUID) bool { return finB9PartySize(g, player) == 4 }

// finB9FullParty and finB9NoFullParty are "if you have a full party" and its
// "instead" complement, read as the ability resolves.
var (
	finB9FullParty = CondBoard{
		Name: "you have a full party",
		Test: func(ctx *Context) bool { return finB9HasFullParty(ctx.Game, ctx.Controller()) },
	}
	finB9NoFullParty = CondBoard{
		Name: "you don't have a full party",
		Test: func(ctx *Context) bool { return !finB9HasFullParty(ctx.Game, ctx.Controller()) },
	}
)

// finB9RunEffect lets a hand-written Effect stand as one step of a Do.
func finB9RunEffect(fn Effect) Applier { return runEffect{fn: fn} }

// finB9CreateTappedCopy is "create a tapped token that's a copy of <card>":
// the copy CreateTokenCopiesTogether makes, entering tapped. `then` runs
// once the token has landed (or at once when the card cannot be copied).
func finB9CreateTappedCopy(ctx *Context, id uuid.UUID, then game.KeyedTail) error {
	tmpl, ok := TokenCopyTemplate(ctx.Game, id)
	if !ok {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: copiedToken(tmpl),
			Count:    1,
			Entry:    game.TokenEntryOptions{Tapped: true},
		}},
		Together: true,
	}, then)
}
