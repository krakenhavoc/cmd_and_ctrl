package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrull Champion — Creature — Thrull {4}{B}, 2/2 (Fallen Empires, S82 OLD
// b854, #5637):
//
//	"Thrull creatures get +1/+1.
//	 {T}: Gain control of target Thrull for as long as you control this
//	 creature."
//
// A layer-7c bonus for every Thrull at the table, this one included, whoever
// controls it (Muscle Sliver's shape). The control change lasts while you
// control the Champion (CR 611.2b), as Mind Flayer's does: a Champion that has
// left, or changed hands, by the time the ability resolves takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "728f0a3f-722d-4936-ade8-31883499f56d",
		Name:         "Thrull Champion",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Thrull"}}, 1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Gain control of target Thrull for as long as you control this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target Thrull", effects.HasSubtype("Thrull")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
					if !ok {
						return nil
					}
					return effects.GainControl{
						Target:     t.ID,
						Controller: item.Controller,
						Duration:   d,
						Label:      "Thrull Champion — control for as long as you control Thrull Champion",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
