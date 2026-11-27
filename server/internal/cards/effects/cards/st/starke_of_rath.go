package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starke of Rath — Legendary Creature — Human Rogue {1}{R}{R}, 2/2 (Tempest,
// S82 OLD, #5628):
//
//	"{T}: Destroy target artifact or creature. That permanent's controller
//	 gains control of Starke. (This effect lasts indefinitely.)"
//
// The controller is read BEFORE the destruction, as Swords to Plowshares reads
// the controller (CR 608.2h): the destroyed permanent has no controller once it
// has gone. Starke changes hands whether or not the permanent was destroyed (an
// indestructible one still takes him), and the control change is indefinite
// (CR 611.2a). A target that became illegal does nothing, Starke included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07bc407e-7ba1-45aa-a562-0597ebe56044",
		Name:         "Starke of Rath",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target artifact or creature. That permanent's controller gains control of Starke.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				controller := card.Controller
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainControl{Target: ctx.Source(), Controller: controller, Duration: game.IndefiniteDuration(),
					Label: "Starke of Rath — gain control"}.Apply(ctx)
			},
		}},
	})
}
