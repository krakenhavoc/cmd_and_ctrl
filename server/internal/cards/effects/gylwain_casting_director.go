package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gylwain, Casting Director — Legendary Creature — Human Bard {1}{G}{W},
// 2/3:
//
//	"Whenever Gylwain or another nontoken creature you control enters,
//	 choose one —
//	 • Create a Royal Role token attached to that creature.
//	 • Create a Sorcerer Role token attached to that creature.
//	 • Create a Monster Role token attached to that creature."
//
// A modal trigger (CR 603.3c) whose three bullets differ only in the
// Role. "That creature" is the entering creature the event names. A
// creature that has left by the time the trigger resolves gets nothing.
// Gylwain's own entry triggers it (the trigger is on the battlefield as
// it enters, CR 603.6a), so he can Role himself.
//
// No simplifications.
func init() {
	role := func(label string, kind RoleKind) game.ModeOption {
		return ModeDoing(label, nil, func(item *game.StackItem, ctx *Context, _ int) error {
			if item.Trigger == nil {
				return nil
			}
			return CreateRoleToken{Role: kind, Host: item.Trigger.Event.CardID}.Apply(ctx)
		})
	}
	gylwain := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.IsCreature() && !c.IsToken()
	}, "Gylwain, Casting Director — choose a Role for the entering creature",
		func(*game.Game, *game.StackItem) error { return nil })
	gylwain.Modes = ChooseOne(
		role("Create a Royal Role token attached to that creature.", RoleRoyal),
		role("Create a Sorcerer Role token attached to that creature.", RoleSorcerer),
		role("Create a Monster Role token attached to that creature.", RoleMonster),
	)
	Register(Spec{
		OracleID:     "e7dfe47f-41fc-4071-9ad3-96cd8d9be52e",
		Name:         "Gylwain, Casting Director",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{gylwain},
	})
}
