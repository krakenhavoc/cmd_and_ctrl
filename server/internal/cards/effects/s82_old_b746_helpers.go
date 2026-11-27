package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b746_helpers.go — shared pieces of S82 OLD batch 746 (#5627).

// TargetsOnlyThisCreature is "target spell or ability that targets only this
// creature" (Silver Wyvern): any spell or ability on the stack whose targets
// are all the clause's own source (game.TargetSpec.ItemTargetsOnlySource).
func TargetsOnlyThisCreature() *game.TargetSpec {
	spec := TargetSpellOrAbility("target spell or ability that targets only this creature")
	spec.ItemTargetsOnlySource = true
	return spec
}

// OldB746DestroyTokensAtNextEndStep is a token creation's keyed tail that
// schedules "destroy it at the beginning of the next end step" over the tokens
// that landed (CR 603.7), as SocB8ExileTokensAtNextEndStep does for an exile
// (Hornet Cannon). An on-disk key: never renamed, never reused.
func OldB746DestroyTokensAtNextEndStep(key, label string) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:         game.StepEnd,
			Controller: ctx.Controller(),
			Label:      label,
			Cards:      append([]uuid.UUID(nil), o.IDs...),
			Body:       oldB746DestroyListedBody,
		}.Apply(ctx)
	})
}

// oldB746DestroyListedBody destroys each listed card still on the battlefield.
// An on-disk key: never renamed, never reused.
var oldB746DestroyListedBody = game.SimpleDelayedBody("destroy/old-b746-listed-cards", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
