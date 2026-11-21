package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veil of Assimilation — Artifact {1}{W} (Phyrexia: All Will Be One):
//
//	"Whenever this artifact or another artifact you control enters, target
//	 creature you control gets +1/+1 and gains vigilance until end of turn."
//
// One trigger per artifact entering under your control, the Veil included.
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dd81e3e-e57a-4b6f-9043-bba0cebb3e90",
		Name:         "Veil of Assimilation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			}, "Veil of Assimilation — target creature you control gets +1/+1 and gains vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Veil of Assimilation — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"}, Label: "Veil of Assimilation — vigilance"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
