package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b614_helpers.go — S82 OLD batch 614 (Eighth Edition, #5607).

// S82OldB614AssignAsThoughUnblocked is "You may have this creature assign its
// combat damage as though it weren't blocked" (Lone Wolf): a layer 6 grant of
// the creature's own flag (CR 510.1c).
func S82OldB614AssignAsThoughUnblocked() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		Label:     "You may have this creature assign its combat damage as though it weren't blocked.",
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.AssignsAsThoughUnblocked = true
		},
	}
}

// S82OldB614SacrificeALand is "sacrifice a land": the controller chooses one
// of the lands they control as the ability resolves (Lesser Gargadon).
func S82OldB614SacrificeALand(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		lands := LandsControlledByPlayer(g, ctx.Controller())
		if len(lands) == 0 {
			return nil
		}
		return SacrificeChoice{Player: ctx.Controller(), Candidates: lands, Question: label}.Apply(ctx)
	}
}

// S82OldB614FoeIsNonblack is the condition of "whenever this creature blocks
// or becomes blocked by a nonblack creature" (Deathgazer): the event names a
// block pair that includes the source, and the other creature is not black.
func S82OldB614FoeIsNonblack(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	foe := SkewerSlingerFoe(ev, source.InstanceID)
	if foe == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(foe)
	return ok && NonBlack()(g, c.Controller, c)
}

// s82OldB614DestroyBody is the delayed half of "destroy that creature at end
// of combat": it destroys the creature the delayed trigger names, if it is
// still the same object (CR 400.7). A registered key (ADR 0041 phase 3).
var s82OldB614DestroyBody = game.SimpleDelayedBody("s82-old-b614/destroy-the-named-creature", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB614DestroyTheFoeAtEndOfCombat schedules the destruction of the other
// creature of the block pair for the beginning of the end of combat step
// (CR 511.2).
func S82OldB614DestroyTheFoeAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		foe := SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
		if foe == uuid.Nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Cards: []uuid.UUID{foe},
			Body:  s82OldB614DestroyBody,
		}.Apply(NewContext(g, item))
	}
}
