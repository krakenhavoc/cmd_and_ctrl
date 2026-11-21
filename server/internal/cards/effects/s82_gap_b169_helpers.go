package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for S82 GAP batch 169 (Jumpstart 2022 enchantments, #5948).

// GapB169TapEnchantedAndInvestigate is Hold for Questioning's entry trigger,
// "tap enchanted permanent and investigate": the Aura taps what it enchants
// (read as the trigger resolves, CR 608.2h) and its controller investigates.
func GapB169TapEnchantedAndInvestigate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if aura, ok := ctx.TriggeringPermanent(); ok && aura.AttachedTo.Kind == game.TargetCard {
		if err := (TapTarget{Target: aura.AttachedTo.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return Investigate{N: 1}.Apply(ctx)
}

// GapB169EnchantedHeldWhileOtherSnow is Winter's Rest's "As long as you
// control another snow permanent, enchanted creature doesn't untap": true for
// the enchanted permanent while the Aura's controller controls a snow
// permanent besides the Aura itself, read through the layers.
func GapB169EnchantedHeldWhileOtherSnow(target *game.Card, g *game.Game, source *game.Card) bool {
	if !AttachedToSource(target, g, source) {
		return false
	}
	for _, id := range PermanentsControlledByMatching(g, source.Controller, Snow()) {
		if id != source.InstanceID {
			return true
		}
	}
	return false
}

// GapB169ControllerGraveyardAtLeast reports whether the controller of the
// enchanted permanent has n or more cards in their graveyard (So Tiny).
func GapB169ControllerGraveyardAtLeast(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !AttachedToSource(target, g, source) {
			return false
		}
		p := g.PlayerByIDForEffect(target.Controller)
		return p != nil && p.Graveyard.Size() >= n
	}
}

// GapB169ControllerGraveyardBelow is the complement of the above, for the
// base "-2/-0" half of So Tiny.
func GapB169ControllerGraveyardBelow(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	atLeast := GapB169ControllerGraveyardAtLeast(n)
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return AttachedToSource(target, g, source) && !atLeast(target, g, source)
	}
}

// GapB169SnowYouControlEntered is "Whenever <this> or another snow permanent
// you control enters": the source itself, or a snow permanent that entered
// under the source's controller.
func GapB169SnowYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSupertype("snow")
}

// GapB169SnowPermanentsAtLeast is the trigger-time half of "if you control
// ten or more snow permanents".
func GapB169SnowPermanentsAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return len(PermanentsControlledByMatching(g, source.Controller, Snow())) >= n
	}
}

// GapB169CondSnowPermanentsAtLeast is its resolution recheck (CR 603.4).
func GapB169CondSnowPermanentsAtLeast(n int) CondBoard {
	return CondBoard{
		Name: "you control ten or more snow permanents",
		Test: func(ctx *Context) bool {
			return len(PermanentsControlledByMatching(ctx.Game, ctx.Controller(), Snow())) >= n
		},
	}
}

// GapB169VehicleIsACreature is Aerial Modification's "As long as enchanted
// permanent is a Vehicle, it's a creature in addition to its other types"
// (layer 4).
func GapB169VehicleIsACreature() game.StaticAbility {
	return game.StaticAbility{
		Label: "As long as enchanted permanent is a Vehicle, it's a creature in addition to its other types.",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && target.HasSubtype("Vehicle")
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !TypeListHasEffects(c.Types, "Creature") {
				c.Types = append(c.Types, "Creature")
			}
		},
	}
}

// GapB169SoldiersEqualToItsPower is Murder Investigation's "create X 1/1
// white Soldier creature tokens, where X is its power": the creature that
// died, by its last-known power (CR 603.10a).
func GapB169SoldiersEqualToItsPower(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
	if !ok || info.Power <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 white Soldier"), N: info.Power}.Apply(NewContext(g, item))
}
