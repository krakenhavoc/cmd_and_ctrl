package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Vampire Lord — Legendary Planeswalker — Sorin {4}{B}{B},
// loyalty 4 (#4077):
//
//	"+1: Up to one target creature gets +2/+0 until end of turn.
//	 −2: Sorin deals 4 damage to any target. You gain 4 life.
//	 −8: Until end of turn, each Vampire you control gains "{T}: Gain
//	     control of target creature.""
//
// The +1 is an "up to one" clause (WithCount(0, 1)): activating it with
// no target is legal and does nothing (CR 601.2c), and a target that left
// in response is skipped (CR 608.2b). The −2's life gain is a separate
// sentence, so a fogged or prevented 4 still gains the 4 (the Sorin
// Markov +2 shape).
//
// THE −8 is a duration grant (ADR 0093, GrantAbilitiesFor): "each Vampire
// you control" is judged ONCE as the ability resolves (CR 611.2c), so a
// Vampire that enters later this turn does not get the ability, and a
// permanent that stops being a Vampire keeps it until end of turn. The
// granted ability is the VAMPIRE's, not Sorin's: its controller
// activates it, {T} taps that Vampire, and CR 302.6 keeps a Vampire that
// came under its controller's control this turn from using it. Its
// effect has no stated duration, so the control it takes is indefinite
// (CR 611.2a), the same shape as Sorin of House Markov's −6.
//
// No simplification.
const sorinVampireLordSteal = "sorin-vampire-lord/gain-control"

// sorinVampireLordRecord files that the -2 had a legal target, read before the damage.
const sorinVampireLordRecord = "sorin-vampire-lord/minus-two"

func init() {
	Register(Spec{
		OracleID:     "753dcc18-3697-49c5-80c6-2d8dc8ee8c9a",
		Name:         "Sorin, Vampire Lord",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Grants: []AbilityGrant{{
			Key: sorinVampireLordSteal,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Gain control of target creature.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect:  sorinVampireLordTakeCreature,
			}},
			Text: "{T}: Gain control of target creature.",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Up to one target creature gets +2/+0 until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: target, Power: 2, Label: "Sorin, Vampire Lord — +2/+0"}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Sorin deals 4 damage to any target. You gain 4 life.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 4)),
				// The life gain waits for a prompt the damage asks (a CR 616 order; CR 608.2c,
				// #4155).
				Effect: Do(
					Instruction(func(ctx *Context) error {
						// CR 608.2b: with its only target gone the ability does
						// nothing at all, life gain included (#4082). A prevented
						// damage still gains the life. That is read here, before
						// the damage can kill the target, and filed for the gain.
						legal := ctx.LegalTargets()
						if len(legal) == 0 {
							return nil
						}
						ctx.RecordStep(sorinVampireLordRecord, nil, 1)
						return DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: 4}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if ctx.StepRecord(sorinVampireLordRecord).N() == 0 {
							return nil
						}
						return GainLife{Amount: 4}.Apply(ctx)
					}),
				),
			},
			{
				Label: "−8: Until end of turn, each Vampire you control gains \"{T}: Gain control of target creature.\"",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GrantAbilitiesFor{
						Match: And(YouControl(), HasSubtype("Vampire")),
						Keys:  []string{sorinVampireLordSteal},
						Label: "Sorin, Vampire Lord — each Vampire you control can gain control of a creature",
					}.Apply(ctx)
				},
			},
		},
	})
}

// sorinVampireLordTakeCreature is the granted ability's effect: its
// controller gains control of the target creature, if it is still a
// legal target.
func sorinVampireLordTakeCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	return GainControl{
		Target:     target,
		Controller: item.Controller,
		Duration:   game.IndefiniteDuration(),
		Label:      "Sorin, Vampire Lord — gain control (no stated duration)",
	}.Apply(ctx)
}
