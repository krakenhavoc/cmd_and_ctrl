package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encouraging Aviator // Jump — Creature — Bird Wizard {2}{U}, 2/3
// // Instant {U} (preparation card, CR 722):
//
//	"Flying
//	 Whenever this creature attacks, it becomes prepared. (While it's
//	 prepared, you may cast a copy of its spell. Doing so unprepares
//	 it.)"
//
//	Jump — "Target creature gains flying until end of turn."
//
// The TRIGGERED "becomes prepared" proof card for ADR 0090: the
// designation arrives from an ability on the stack rather than from
// the entry, through the BecomePrepared primitive. An Aviator that is
// already prepared when it attacks gains nothing (CR 722.3a — "a
// permanent can't gain this designation if the permanent already has
// it"), so attacking twice without casting Jump leaves exactly one
// copy in exile. An Aviator removed in response is not there to
// become prepared, and the ability does nothing.
//
// No simplification.
const EncouragingAviatorOracleID = "d6accaed-ff35-4324-b31b-35e6837bc079"

func init() {
	effects.Register(effects.Spec{
		OracleID:        EncouragingAviatorOracleID,
		Name:            "Encouraging Aviator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Encouraging Aviator — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     EncouragingAviatorOracleID + "#1",
		Name:         "Jump",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"flying"},
					Label:    "Jump — flying until end of turn",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
