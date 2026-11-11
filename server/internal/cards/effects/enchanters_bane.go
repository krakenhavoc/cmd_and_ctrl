package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enchanter's Bane — Enchantment {1}{R} (Duskmourn Commander, #5368):
//
//	"At the beginning of your end step, target enchantment deals damage
//	 equal to its mana value to its controller unless that player
//	 sacrifices it."
//
// A targeted end-step trigger. At resolution the target's controller is
// asked, in the shape of the punisher (PickOption, ADR 0100): take the
// damage (always offered, so it is first) or sacrifice the enchantment.
// The damage's source is the enchantment itself, with the mana value it
// has as the choice is made; a target that left or stopped being an
// enchantment by resolution is illegal and the ability does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f0548f8-5dec-44b5-b5a9-e55fa4366e08",
		Name:         "Enchanter's Bane",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventBeginEndStep, ByYou,
				"Enchanter's Bane — target enchantment deals damage equal to its mana value to its controller unless that player sacrifices it",
				enchantersBaneResolve),
				TargetPermanent("target enchantment", Enchantment())),
		},
	})
}

const enchantersBaneSacrificeLabel = "Sacrifice the enchantment"

func enchantersBaneResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	return PickOption{
		Player:   c.Controller,
		Question: "Enchanter's Bane — take damage equal to " + c.Name + "'s mana value, or sacrifice it",
		Options: []game.ChoiceOption{
			{Label: "Take damage equal to its mana value"},
			{Label: enchantersBaneSacrificeLabel},
		},
		ThenKey: enchantersBaneAnswer,
		Carry:   []uuid.UUID{t.ID},
	}.Apply(ctx)
}

// enchantersBaneAnswer is the controller's answer: sacrifice the
// enchantment, or take its mana value in damage from it. An on-disk key:
// never renamed, never reused.
var enchantersBaneAnswer = OptionPickThen("option-pick/enchanters-bane-answer", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	enchantment := r.Carry[0]
	if r.Option != nil && r.Option.Label == enchantersBaneSacrificeLabel {
		return nothingIfGone(ctx.Game.SacrificePermanentForEffect(enchantment))
	}
	c, ok := ctx.Game.LookupCardForEffect(enchantment)
	if !ok {
		return nil
	}
	return DealDamage{Source: enchantment, Target: r.Chooser, Amount: c.ManaValue()}.Apply(ctx)
})
