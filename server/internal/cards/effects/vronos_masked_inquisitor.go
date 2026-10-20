package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vronos, Masked Inquisitor — Legendary Planeswalker — Vronos
// {3}{U}{U}, loyalty 5:
//
//	"+1: Up to two other target planeswalkers you control phase out at
//	     the beginning of the next end step. (Treat them and anything
//	     attached to them as though they don't exist until your next
//	     turn.)
//	 −2: For each opponent, return up to one target nonland permanent
//	     that player controls to its owner's hand.
//	 −7: Target artifact you control becomes a 9/9 Construct artifact
//	     creature and gains vigilance, indestructible, and "This creature
//	     can't be blocked.""
//
// +1 targets now and phases out later: the targets are checked as the
// ability resolves (CR 608.2b), and each still-legal one is named as an
// OBJECT on a CR 603.7 delayed trigger for the next end step. That
// trigger phases out each named object still on the battlefield
// (CR 702.26), with whatever is attached to it; a walker that left and
// came back in between is a new object and stays (CR 400.7). They phase
// in during their controller's next untap step, which is Vronos's
// controller's next turn, as the reminder says.
//
// −2 is The Theorist, Jace Beleren's −2 with "nonland permanent": one
// target per opponent at most (EachDifferentController), each returned
// if still legal.
//
// −7 has no stated duration (CR 611.2a), so the change lasts as long as
// the artifact is that object. It becomes a creature with base 9/9 and
// the Construct subtype in addition to its other types (CR 205.1b),
// gains vigilance and indestructible, and can't be blocked (the
// Creeping Tar Pit restriction).
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "86851bdb-9b2f-4407-97a1-a6a9f2e0ef66",
		Name:         "Vronos, Masked Inquisitor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Up to two other target planeswalkers you control phase out at the beginning of the next end step.",
				Cost:  LoyaltyCost(1),
				Targets: Another(TargetPermanent("up to two other target planeswalkers you control",
					Planeswalker(), YouControl())).WithCount(0, 2),
				Effect: vronosPlusOne,
			},
			{
				Label: "−2: For each opponent, return up to one target nonland permanent that player controls to its owner's hand.",
				Cost:  LoyaltyCost(-2),
				Targets: TargetPermanent("up to one target nonland permanent each opponent controls",
					Nonland(), OpponentControls()).
					WithCount(0, 0).EachDifferent(EachDifferentController()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b35BounceChosenPermanents(NewContext(g, item))
				},
			},
			{
				Label:   "−7: Target artifact you control becomes a 9/9 Construct artifact creature and gains vigilance, indestructible, and \"This creature can't be blocked.\"",
				Cost:    LoyaltyCost(-7),
				Targets: TargetPermanent("target artifact you control", Artifact(), YouControl()),
				Effect:  vronosMinusSeven,
			},
		},
	})
}

// vronosPhaseOutObjectsBody is the +1's delayed trigger: each named
// object still on the battlefield phases out.
var vronosPhaseOutObjectsBody = game.DelayedBody("vronos/phase-out-the-objects",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		var ids []uuid.UUID
		for _, ref := range p.Objects {
			if !g.Battlefield.Contains(ref.ID) {
				continue
			}
			if c, ok := g.LookupCardForEffect(ref.ID); ok && c.ObjectEpoch == ref.Epoch {
				ids = append(ids, ref.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return g.PhaseOutForEffect(item.SourceCardID, ids...)
	})

func vronosPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	var objects []game.ObjectRef
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		ref, ok := g.PermanentRefForEffect(t.ID)
		if !ok || !g.Battlefield.Contains(t.ID) {
			continue
		}
		ids = append(ids, t.ID)
		objects = append(objects, ref)
	}
	if len(objects) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Vronos, Masked Inquisitor — the chosen planeswalkers phase out",
		Cards:  ids,
		Body:   vronosPhaseOutObjectsBody,
		Params: game.EffectParams{Objects: objects},
	}.Apply(ctx)
}

func vronosMinusSeven(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !g.Battlefield.Contains(t.ID) {
			continue
		}
		return ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.AddTypesMod("Artifact", "Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Construct")...),
				game.SetBasePowerMod(9),
				game.SetBaseToughnessMod(9),
				game.AddKeywordsMod("vigilance", "indestructible"),
				game.AddRestrictionsMod(game.CantBeBlocked),
			},
			Duration: game.IndefiniteDuration(),
			Label:    "Vronos, Masked Inquisitor — a 9/9 Construct artifact creature with vigilance, indestructible, and \"This creature can't be blocked.\"",
		}.Apply(ctx)
	}
	return nil
}
