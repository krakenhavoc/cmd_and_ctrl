package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Varchild's Crusader — Creature — Human Knight {3}{R}, 3/2 (Alliances,
// #5634):
//
//	"{0}: This creature can't be blocked this turn except by Walls. Sacrifice
//	 this creature at the beginning of the next end step."
//
// One ability, two sentences: the evasion is a until-end-of-turn block rule
// pinned to this object (Gingerbrute's shape) and the sacrifice a delayed
// trigger for the next end step, taken only while it is still this permanent
// (CR 400.7). Activated before blockers are declared it matters; the sacrifice
// happens whenever it is activated.
//
// No simplification.
func init() {
	const name = "Varchild's Crusader"
	effects.Register(effects.Spec{
		OracleID:     "45351d9d-453f-4245-84c0-053eb8aca345",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: This creature can't be blocked this turn except by Walls. Sacrifice this creature at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{0}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := (effects.CantBeBlockedThisTurnExceptBy{
					Target:   item.SourceCardID,
					Subtypes: []string{ctx.TextSubtype("Wall")},
					Text:     "Walls",
					Label:    name + " — can't be blocked except by Walls",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.OldB507SacrificeThisAtTheNextEndStep(name+" — sacrifice this creature")(g, item)
			},
		}},
	})
}
