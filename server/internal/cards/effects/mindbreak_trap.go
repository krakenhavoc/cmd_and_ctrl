package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindbreak Trap — Instant — Trap {2}{U}{U} (EDHREC rank 550):
//
//	"If an opponent cast three or more spells this turn, you may pay
//	 {0} rather than pay this spell's mana cost.
//	 Exile any number of target spells."
//
// The trap clause is an alternative cost (CR 118.9) with Refraction
// Trap's shape: a Condition read off each opponent's per-turn cast
// tally, so it is offered only once one opponent has cast three spells
// this turn (copies are not cast and do not count, CR 707.10).
//
// The spells are exiled, not countered (ExileTargetSpell, #1318): a
// spell that can't be countered is exiled all the same, and nothing
// watching for a counter fires. Each still-legal target is exiled in
// turn, the next once the last has settled, so a commander spell whose
// owner is asked about the command zone (CR 903.9) holds the rest
// until they answer. A target that left the stack in response is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48712d36-5c13-4ded-87a4-68106da4a5cd",
		Name:         "Mindbreak Trap",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("any number of target spells").WithCount(0, 0),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent cast three or more spells this turn, pay {0}",
			ManaCost:  "{0}",
			Condition: anOpponentCastThreeOrMoreSpellsThisTurn,
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var spells []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					spells = append(spells, t.ID)
				}
			}
			return mindbreakExileEach(ctx, spells)
		},
	})
}

// anOpponentCastThreeOrMoreSpellsThisTurn is the trap's condition.
func anOpponentCastThreeOrMoreSpellsThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		if g.CastTallyFor(p.ID).Total >= 3 {
			return true
		}
	}
	return false
}

// mindbreakExileEach exiles the first spell and hands the rest to the
// exile's continuation.
func mindbreakExileEach(ctx *Context, spells []uuid.UUID) error {
	if len(spells) == 0 {
		return nil
	}
	rest := spells[1:]
	return ExileTargetSpell{StackID: spells[0], Then: func(ctx *Context, _ bool) error {
		return mindbreakExileEach(ctx, rest)
	}}.Apply(ctx)
}
