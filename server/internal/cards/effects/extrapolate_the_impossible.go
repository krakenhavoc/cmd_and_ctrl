package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extrapolate the Impossible — Sorcery {1}{B}:
//
//	"You may reveal exactly two cards you own with different names from
//	 outside the game. An opponent chooses one of them. You put that
//	 card into your hand."
//
// Scryfall rulings (2026-08-21): "The other card stays outside the
// game." and "In a casual game, a card you choose from outside the game
// comes from your personal collection. In a tournament event, a card you
// choose from outside the game must come from your sideboard." A deck's
// sideboard is that pool here (ADR 0191), so the card works whenever the
// deck brought one.
//
// Three beats, each a prompt the engine already knows how to ask:
//
//   - the caster chooses exactly two cards with different names from
//     their sideboard, or none ("you MAY reveal"): a choose_cards over
//     the container with a set-level Validate, so a single card or a
//     pair of the same name is refused, and the enumerator offers only
//     the pairs the engine would accept;
//   - both are revealed to the table, then the caster names the opponent
//     who chooses (the engine skips that question at two seats), and
//     that opponent picks one of the two they were just shown (a
//     reveal_pick, whose candidates are public because they were
//     revealed);
//   - the chosen card goes to the caster's hand. The other stays in the
//     sideboard, and stays known to the table, which saw it.
//
// A sideboard with fewer than two differently named cards offers no
// prompt and does nothing. Nothing is shipped stronger than printed.
func init() {
	Register(Spec{
		OracleID:     "1f47c869-f4a9-43b0-aec0-8da4e3a3fec5",
		Name:         "Extrapolate the Impossible",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return extrapolateResolve(item, ctx)
		},
	})
}

// extrapolateResolve asks the caster for the two cards, if the sideboard
// can supply two with different names.
//
// Caller holds g.mu.
func extrapolateResolve(item *game.StackItem, ctx *Context) error {
	// A deck with no sideboard chooses from the card database (ADR
	// 0198), which always holds two cards with different names; the
	// prompt's own set rule is what keeps the answer to two.
	if ctx.Game.HasSideboardPool(ctx.Controller()) {
		var pool []game.Card
		for _, id := range ctx.Game.OutsideCandidatesForEffect(ctx.Controller(), nil) {
			if c, ok := ctx.Game.LookupOutsideCardForEffect(id); ok {
				pool = append(pool, c)
			}
		}
		if !atLeastTwoNames(pool) {
			return nil
		}
	}
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Extrapolate the Impossible — reveal exactly two cards with different names from outside the game, or none",
		Min:      0,
		Max:      2,
		Validate: exactlyTwoDifferentNames,
		Then:     extrapolateReveal(item),
	})
	return nil
}

// extrapolateReveal reveals the pair and hands them to an opponent.
//
// A package-level constructor closing over the item alone, the
// StackItem.Effect contract.
func extrapolateReveal(item *game.StackItem) func(g *game.Game, picked []uuid.UUID) error {
	return func(g *game.Game, picked []uuid.UUID) error {
		if len(picked) != 2 {
			return nil
		}
		ctx := NewContext(g, item)
		g.RevealOutsideForEffect(game.RevealSpec{
			Player: ctx.Controller(),
			Source: ctx.Source(),
			Reason: "Extrapolate the Impossible — reveal two cards from outside the game",
			Cards:  picked,
		})
		return ChoosePlayer{
			Among:    Opponents,
			Question: "Extrapolate the Impossible — choose an opponent to pick one of the revealed cards",
			Then:     extrapolateChoose(picked),
		}.Apply(ctx)
	}
}

// extrapolateChoose is "an opponent chooses one of them. You put that
// card into your hand."
//
// Caller holds g.mu.
func extrapolateChoose(revealed []uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		chooser := ctx.ChosenPlayer()
		if chooser == uuid.Nil {
			// Every opponent has left (CR 800.4a): nobody can choose,
			// and the card stays outside the game.
			return nil
		}
		return RevealPick{
			Player:   chooser,
			Owner:    ctx.Controller(),
			Question: "Extrapolate the Impossible — choose one of these cards; it goes into their hand and the other stays outside the game",
			Cards:    revealed,
			Min:      1,
			Max:      1,
			Then:     extrapolateTake,
		}.Apply(ctx)
	}
}

// extrapolateTake puts the chosen card into the caster's hand. The card
// not chosen stays where it was.
//
// Caller holds g.mu.
func extrapolateTake(ctx *Context, picked, _ []uuid.UUID) error {
	for _, id := range picked {
		if err := ctx.Game.BringFromOutsideForEffect(ctx.Controller(), id); err != nil {
			return err
		}
	}
	return nil
}

// exactlyTwoDifferentNames is the set-level rule of the first prompt:
// "exactly two cards you own with different names".
func exactlyTwoDifferentNames(picked []game.Card) bool {
	return len(picked) == 2 && DifferentNames(picked)
}

// atLeastTwoNames reports whether a pool holds two cards with different
// names, which is the least that can answer the first prompt.
func atLeastTwoNames(pool []game.Card) bool {
	for i := range pool {
		for j := i + 1; j < len(pool); j++ {
			if DifferentNames([]game.Card{pool[i], pool[j]}) {
				return true
			}
		}
	}
	return false
}
