package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contractual Safeguard — Instant {2}{W} (New Capenna Commander, #5491):
//
//	"Addendum — If you cast this spell during your main phase, put a shield
//	 counter on a creature you control. (If it would be dealt damage or
//	 destroyed, remove a shield counter from it instead.)
//	 Choose a kind of counter on a creature you control. Put a counter of
//	 that kind on each other creature you control."
//
// Unbreakable Formation's addendum, read as the spell resolves (the stack
// has to empty for a step to end, so that is where it was cast). The shield
// goes on a creature chosen then, not targeted, and the kind is chosen after
// it, so the new shield counter may be the kind chosen. The kind is picked
// together with the creature it is on, because "each other creature" is
// every creature but that one: each (creature, kind) pair is an option,
// creatures in battlefield order and kinds in name order. Nothing is chosen
// and nothing is put when no creature has a counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ac43884-9f03-4945-98ea-c744d0403f32",
		Name:         "Contractual Safeguard",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b14InYourMainPhase(ctx.Game, ctx.Controller()) {
				return contractualSafeguardShield(ctx)
			}
			return contractualSafeguardKind(ctx)
		},
	})
}

// contractualSafeguardShield is the addendum's pick of the creature.
func contractualSafeguardShield(ctx *Context) error {
	var options []game.ChoiceOption
	var creatures []uuid.UUID
	for _, c := range sncB07PermanentsYouControl(ctx.Game, ctx.Controller(), true) {
		options = append(options, game.ChoiceOption{Label: c.Name})
		creatures = append(creatures, c.InstanceID)
	}
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), creatures[0], game.CounterShield, 1); err != nil {
			return err
		}
		return contractualSafeguardKind(ctx)
	}
	return PickOption{
		Question: "Contractual Safeguard — put a shield counter on which creature?",
		Options:  options,
		ThenKey:  contractualSafeguardShielded,
		Carry:    creatures,
	}.Apply(ctx)
}

// contractualSafeguardShielded puts the shield counter, then goes on to the
// kind. An on-disk key.
var contractualSafeguardShielded = OptionPickThen("option-pick/contractual-safeguard-shield", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(r.Carry) {
		return nil
	}
	if onBattlefield(ctx.Game, r.Carry[r.Index]) {
		if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), r.Carry[r.Index], game.CounterShield, 1); err != nil {
			return err
		}
	}
	return contractualSafeguardKind(ctx)
})

// contractualSafeguardKind is the choice of a kind on a creature.
func contractualSafeguardKind(ctx *Context) error {
	options, creatures, kinds := sncB07CreaturePairs(ctx.Game, ctx.Controller())
	switch len(options) {
	case 0:
		return nil
	case 1:
		return contractualSafeguardSpread(ctx, creatures[0], kinds[0])
	}
	return PickOption{
		Question: "Contractual Safeguard — choose a kind of counter on a creature you control",
		Options:  options,
		ThenKey:  contractualSafeguardChosen,
		Carry:    creatures,
		Words:    kinds,
	}.Apply(ctx)
}

// contractualSafeguardChosen is the chosen pair, keyed. An on-disk key.
var contractualSafeguardChosen = OptionPickThen("option-pick/contractual-safeguard-kind", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(r.Carry) || r.Index >= len(r.Words) {
		return nil
	}
	return contractualSafeguardSpread(ctx, r.Carry[r.Index], r.Words[r.Index])
})

// contractualSafeguardSpread puts a counter of the kind on each creature the
// controller has other than the chosen one.
func contractualSafeguardSpread(ctx *Context, chosen uuid.UUID, kind string) error {
	for _, c := range sncB07PermanentsYouControl(ctx.Game, ctx.Controller(), true) {
		if c.InstanceID == chosen {
			continue
		}
		if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), c.InstanceID, kind, 1); err != nil {
			return err
		}
	}
	return nil
}
