package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func doesntUntapDuringYourUntapStep() game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label: "doesn't untap during your untap step",
		Restricts: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target != nil && source != nil && target.InstanceID == source.InstanceID
		},
	}
}

func enchantedDoesntUntap() game.UntapStepRestriction {
	return game.UntapStepRestriction{Label: "enchanted creature doesn't untap", Restricts: AttachedToSource}
}

func doesntUntapDuringTheirControllersUntapSteps(match CardPredicate) game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label: "doesn't untap during its controller's untap step",
		Restricts: func(target *game.Card, g *game.Game, _ *game.Card) bool {
			return target != nil && (match == nil || match(g, target.Controller, *target))
		},
	}
}

// DoesntUntapNextUntapStep records the fixed set from a resolved one-shot
// effect. Player uuid.Nil means each permanent's current controller.
type DoesntUntapNextUntapStep struct {
	Targets []uuid.UUID
	Player  uuid.UUID
	Label   string
}

func (d DoesntUntapNextUntapStep) Apply(ctx *Context) error {
	for _, id := range d.Targets {
		if err := ctx.Game.SkipNextUntapForEffect(id, d.Player); err != nil {
			return err
		}
	}
	return nil
}

// TapAndFreeze implements the common "tap ...; those permanents don't
// untap during ... next untap step" resolution.
type TapAndFreeze struct {
	Targets []uuid.UUID
	Player  uuid.UUID
	Label   string
}

func (t TapAndFreeze) Apply(ctx *Context) error {
	for _, id := range t.Targets {
		if err := ctx.Game.TapTargetForEffect(id); err != nil {
			return err
		}
	}
	return (DoesntUntapNextUntapStep{Targets: t.Targets, Player: t.Player, Label: t.Label}).Apply(ctx)
}
