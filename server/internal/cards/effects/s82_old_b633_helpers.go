package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82OldB633ControlsColor is "you control a <colour> permanent" for the
// trigger's controller, read now.
func s82OldB633ControlsColor(ctx *Context, letter string) bool {
	return B18ControlsPermanentOfColor(ctx.Game, ctx.Controller(), ctx.TextColor(letter))
}

// S82OldB633RakaSanctuary is Raka Sanctuary's upkeep trigger: "if you control
// a white or blue permanent, this enchantment deals 1 damage to target
// creature. If you control a white permanent and a blue permanent, it deals 3
// damage instead." The intervening if (CR 603.4) is checked as it triggers and
// again on resolution; the "instead" is read as it resolves (CR 608.2c).
func S82OldB633RakaSanctuary(label string) game.TriggeredAbility {
	either := CondBoard{
		Name: "you control a white or blue permanent",
		Test: func(ctx *Context) bool {
			return s82OldB633ControlsColor(ctx, "W") || s82OldB633ControlsColor(ctx, "U")
		},
	}
	t := On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller &&
			(B18ControlsPermanentOfColor(g, source.Controller, source.TextColor("W")) ||
				B18ControlsPermanentOfColor(g, source.Controller, source.TextColor("U")))
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		n := 1
		if s82OldB633ControlsColor(ctx, "W") && s82OldB633ControlsColor(ctx, "U") {
			n = 3
		}
		return DealDamageToTheTarget(n)(g, item)
	})
	return InterveningIf(either, Targeting(t, TargetCreature("target creature")))
}

// S82OldB633MaskOfIntolerance is Mask of Intolerance's trigger: "At the
// beginning of each player's upkeep, if there are four or more basic land
// types among lands that player controls, this artifact deals 3 damage to that
// player." The player is the one whose upkeep it is, and the count is theirs
// (domain, CR 305.6) as it triggers and again on resolution (CR 603.4).
func S82OldB633MaskOfIntolerance(label string) game.TriggeredAbility {
	return On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor != uuid.Nil && DomainCount(g, ev.Actor) >= 4
	}, label, func(g *game.Game, item *game.StackItem) error {
		player := TriggeringActor(item)
		if player == uuid.Nil || DomainCount(g, player) < 4 {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: player, Amount: 3}.Apply(NewContext(g, item))
	})
}
