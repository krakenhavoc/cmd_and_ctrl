package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b755_helpers.go — Tempest, batch 755 (S82 older sets, #5628).

// OldB755DestroyAtNextEndStep is "Destroy that creature at the beginning of
// the next end step": a delayed trigger (CR 603.7) that destroys the object
// as it is now, if it is still that object on the battlefield (CR 400.7,
// 603.7c). It reuses Stone Giant's delayed body.
func OldB755DestroyAtNextEndStep(ctx *Context, id uuid.UUID, label string) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  label + " — destroy " + c.Name,
		Cards:  []uuid.UUID{id},
		Body:   oldB499DestroyTheObjectBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}

// OldB755TargetGetsThenDestroyAtEndStep is "Target creature gets +P/+0 [and
// gains <keywords>] until end of turn. Destroy that creature at the beginning
// of the next end step" (Mogg Cannon, Blood Frenzy): the pump and the grant
// run on the target as it resolves, then the delayed destruction is
// scheduled. A target that is no longer legal takes the whole ability with it
// (CR 608.2b).
func OldB755TargetGetsThenDestroyAtEndStep(label string, power int, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Label: label}).Apply(ctx); err != nil {
			return err
		}
		if len(keywords) > 0 {
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: label + " — " + keywords[0]}).Apply(ctx); err != nil {
				return err
			}
		}
		return OldB755DestroyAtNextEndStep(ctx, id, label)
	}
}

// OldB755CreaturesOnTheBattlefieldOtherThan counts "each other creature on
// the battlefield" (Mogg Squad): every creature of every player but the
// source, read through the layers on each pass.
func OldB755CreaturesOnTheBattlefieldOtherThan(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != source.InstanceID && c.IsCreature() {
			n++
		}
	}
	return n
}

// OldB755BeforeTheCombatDamageStep is Blood Frenzy's cast window: any time
// before the first combat damage step of the turn (CR 506.8), the first strike
// damage step included. A turn whose combat damage steps are skipped keeps the
// window to the end of the declare attackers step (CR 506.8e), which the step
// order gives for free: the later steps are never reached.
func OldB755BeforeTheCombatDamageStep(g *game.Game, _ uuid.UUID, _ game.Card) bool {
	switch game.PhaseOf(g.Turn.Step) {
	case game.PhaseBeginning, game.PhasePrecombatMain:
		return true
	case game.PhaseCombat:
		switch g.Turn.Step {
		case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers:
			return true
		}
	}
	return false
}
