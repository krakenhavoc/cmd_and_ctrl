package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b756_helpers.go — Tempest, batch 756 (S82 older sets, #5628).

// OldB756ThisGetsAndGainsThenDestroyAtEndStep is "This creature gets +P/+0 and
// gains <keyword> until end of turn. Destroy this creature at the beginning of
// the next end step" (Crazed Armodon): the pump and the grant on the source as
// it is now, then a delayed trigger (CR 603.7) that destroys that object at the
// next end step if it is still on the battlefield (CR 400.7, 603.7c). It reuses
// Stone Giant's delayed destroy body.
func OldB756ThisGetsAndGainsThenDestroyAtEndStep(label string, power int, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := item.SourceCardID
		c, ok := g.LookupCardForEffect(id)
		if !ok || !OnBattlefield(g, id) {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Label: label}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label + " — " + keyword}).Apply(ctx); err != nil {
			return err
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label + " — destroy " + c.Name,
			Cards:  []uuid.UUID{id},
			Body:   oldB499DestroyTheObjectBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}.Apply(ctx)
	}
}
