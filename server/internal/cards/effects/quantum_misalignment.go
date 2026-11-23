package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quantum Misalignment — Sorcery {4}{U}:
//
//	"Create a token that's a copy of target creature you control,
//	 except it isn't legendary.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The copy's exception is Miirym's: the Legendary supertype comes off
// the copied type line (CR 707.9b), so a legendary creature's copy
// does not fall to the legend rule. Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "197582b9-4c86-41a3-ad0c-789a0db8e087",
		Name:            "Quantum Misalignment",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature you control", YouControl()),
		OnResolve:       NonLegendaryTokenCopyOfTheTarget,
	})
}

// nonLegendaryTokenCopyOfTheTarget is "Create a token that's a copy of target
// creature you control, except it isn't legendary" (Quantum Misalignment,
// Multiversal Recruitment): the copy's Legendary supertype comes off the
// copied type line (CR 707.9b).
func NonLegendaryTokenCopyOfTheTarget(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       id,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = RemoveLegendaryFromTypeLine(t.TypeLine)
		},
	}.Apply(ctx)
}
