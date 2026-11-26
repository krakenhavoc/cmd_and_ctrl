package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// View from Above — Instant {1}{U} (Conflux):
//
//	"Target creature gains flying until end of turn. If you control a white
//	 permanent, return View from Above to its owner's hand."
//
// The grant is the first sentence; the condition is read as the spell
// resolves, through the layered colours of the permanents you control. The
// return is the spell's own resolution exit (it goes to its owner's hand
// instead of the graveyard, CR 608.2n), so it is not a counter and a copy of
// the spell, which is not a card, just ceases to exist (CR 707.10a). With the
// target illegal the spell does nothing and is countered by the rules, so it
// goes to the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a128587-c751-4b27-b608-d4e47646b80e",
		Name:         "View from Above",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      viewFromAbovePurpose(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id := effects.FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "View from Above — flying"}).Apply(ctx); err != nil {
					return err
				}
			}
			white := ctx.TextColor("W")
			if effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasColor(white) }) > 0 {
				ctx.Game.ReturnResolvingSpellToHandForEffect(item.ID)
			}
			return nil
		},
	})
}

// viewFromAbovePurpose declares the flying grant on the target (#4312) beside
// the answer it is.
func viewFromAbovePurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))
	p.Answers = game.AnswerCombatGrant
	return p
}
