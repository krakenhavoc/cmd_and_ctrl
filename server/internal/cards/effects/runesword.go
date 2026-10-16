package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runesword — Artifact {6}:
//
//	"{3}, {T}: Target attacking creature gets +2/+0 until end of turn.
//	 When that creature leaves the battlefield this turn, sacrifice this
//	 artifact. If the creature deals damage to a creature this turn, the
//	 creature dealt damage can't be regenerated this turn. If a creature
//	 dealt damage by the targeted creature would die this turn, exile
//	 that creature instead."
//
// Four effects on the one target object, in printed order:
//
//   - the +2/+0 until end of turn;
//   - "When that creature leaves the battlefield this turn, sacrifice
//     this artifact" is an event-delayed trigger (CR 603.7) on the
//     object's leaving, pinned to it, whose body sacrifices Runesword if
//     it is still the object that made it;
//   - the two riders are ScopeDealtDamageBy records naming the creature
//     (#1946): read live against the turn's record of which objects it
//     damaged, so they reach every creature it damages, combat or not,
//     and they last until cleanup even if the creature leaves.
//
// "If the creature deals damage to a creature this turn" looks forward
// from the resolution, so the can't-be-regenerated rider counts only the
// damage dealt after it (Mod.DamageSince). "A creature dealt damage by
// the targeted creature" is not limited that way, so the exile counts
// the whole turn's damage, first-strike damage dealt before a
// regular-damage activation included. A creature damaged and then
// flickered is a new object (CR 400.7).
//
// No simplification.
func init() {
	runeswordSacrificeBody = game.SimpleDelayedBody("ltb/sacrifice-the-source", sacrificeDelayedTriggerSource)
	Register(Spec{
		OracleID:     "964c2003-cef0-4ac7-9f66-4e20893c8e50",
		Name:         "Runesword",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target attacking creature gets +2/+0 until end of turn. When that creature leaves the battlefield this turn, sacrifice this artifact. If the creature deals damage to a creature this turn, the creature dealt damage can't be regenerated this turn. If a creature dealt damage by the targeted creature would die this turn, exile that creature instead.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect:  runeswordResolve,
		}},
	})
}

// runeswordSacrificeBody is the delayed trigger's body: sacrifice the
// trigger's source (Runesword). A registered key (ADR 0041 phase 3), so
// a table with the trigger waiting is a restore point. Assigned in init:
// a var initialiser would be an initialisation cycle through the exit
// primitives (dorotheaSacrificeBody's reason).
var runeswordSacrificeBody game.BodyRef

func runeswordResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Runesword — +2/+0"}).Apply(ctx); err != nil {
		return err
	}
	d := g.PinnedTo(g.UntilEndOfTurnDuration(), id)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Runesword — sacrifice it",
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectLeftCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         runeswordSacrificeBody,
	})
	g.CantBeRegeneratedIfDealtDamageByForEffect(ctx.Source(), ref, true,
		"Runesword — creatures it damages can't be regenerated this turn")
	g.ExileIfDealtDamageByWouldDieThisTurnForEffect(ctx.Source(), ref, uuid.Nil, false,
		"Runesword — exile creatures it damaged instead of dying")
	return nil
}

// theListedObjectLeft is the condition of "when that creature leaves the
// battlefield": an EventLTB of the listed object, to any zone.
func theListedObjectLeft(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventLTB && len(dt.Cards) > 0 && ev.CardID == dt.Cards[0]
}
