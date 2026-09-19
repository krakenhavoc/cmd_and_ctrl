package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Shattered Pact — Enchantment — Case, {2}:
//
//	When this Case enters, search your library for a basic land card,
//	reveal it, put it into your hand, then shuffle.
//	To solve — There are five colors among permanents you control.
//	(If unsolved, solve at the beginning of your end step.)
//	Solved — At the beginning of combat on your turn, target creature
//	you control gains flying, double strike, and vigilance until end
//	of turn.
//
// The first Case in the catalog (ADR 0071). All three lines work, and
// the interesting one is the middle:
//
//	To solve — …
//
// is not a keyword and not machinery. CR 719.3a spells it out as a
// triggered ability with an intervening if — "at the beginning of
// your end step, if [condition], this Case becomes solved" — so the
// whole clause is the ToSolve constructor plus this file's
// fiveColorsAmongYourPermanents. The condition is checked when the
// trigger would go on the stack AND again as it resolves (CR 603.4),
// which matters here: a Case whose fifth colour is killed in response
// does not solve.
//
// The solved line is an ORDINARY triggered ability with a gate. There
// is no "solved trigger" type: WhenSolved stamps CaseSolved() onto a
// trigger built by the same constructors every other card uses, and
// below that gate the harvester cannot see it at all. Before the Case
// is solved, its combat trigger does not fire, is not prompted, and
// does not ask anybody to pick a target — which is the difference
// between a gate and an `if solved` inside AppliesTo, because the
// second would still have chosen a target.
func init() {
	Register(Spec{
		OracleID:     "d00a089b-7d0d-4b89-902b-7b914e892178",
		Name:         "Case of the Shattered Pact",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Case of the Shattered Pact — search for a basic land",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Case of the Shattered Pact — search for a basic land card",
					}.Apply(NewContext(g, item))
				}),
			ToSolve("Case of the Shattered Pact — to solve: five colors among permanents you control",
				fiveColorsAmongYourPermanents),
			WhenSolved(Targeting(
				On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
					"Case of the Shattered Pact — solved: target creature gains flying, double strike and vigilance",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						target, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{
							Target:   target.ID,
							Keywords: []string{"flying", "double strike", "vigilance"},
							Label:    "Case of the Shattered Pact — flying, double strike and vigilance",
						}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl()),
			)),
		},
	})
}

// fiveColorsAmongYourPermanents is the Case's solve condition:
// "There are five colors among permanents you control."
//
// It counts colours across the whole board the player controls, not
// per permanent, so one five-colour commander solves it alone and so
// do five mono-coloured creatures. Colourless permanents contribute
// nothing and are not a sixth colour (CR 105.1).
//
// Reads EFFECTIVE colours, through the same Card.HasColor the
// targeting predicates use, so a layer-5 colour change counts — which
// is what "there are five colors among permanents you control" says
// rather than "five colours are printed".
//
// Runs under g.mu: read-only, *ForEffect accessors only.
func fiveColorsAmongYourPermanents(g *game.Game, controller, _ uuid.UUID) bool {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		for _, color := range []string{"W", "U", "B", "R", "G"} {
			if c.HasColor(color) {
				seen[color] = true
			}
		}
	}
	return len(seen) == 5
}
