package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// silentHallcreeperLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen" memory is kept under.
const silentHallcreeperLabel = "Silent Hallcreeper — choose one"

// Silent Hallcreeper — Enchantment Creature — Horror {1}{U}, 1/1
// (slice 296-m):
//
//	"This creature can't be blocked.
//	 Whenever this creature deals combat damage to a player, choose
//	 one that hasn't been chosen —
//	 • Put two +1/+1 counters on this creature.
//	 • Draw a card.
//	 • This creature becomes a copy of another target creature you
//	   control."
//
// The evasion is a plain RestrictSelf(CantBeBlocked). The damage
// trigger is combatDamageToPlayerBy's own shape, with a real mode
// choice (TriggeredAbility.Modes, #764) answered through a mode_pick
// prompt.
//
// "That hasn't been chosen" is ChooseOneNotChosen (ADR 0097, #1749):
// the engine remembers, on this OBJECT, which bullets its ability has
// chosen, for as long as the object exists. The 2024-09-20 ruling
// says exactly that — a Hallcreeper that leaves the battlefield and
// returns "will be a new object with no memory of the modes that were
// chosen" — and MoveCard's CR 400.7 forgetting is what delivers it. A
// fourth trigger, with all three used, is removed with no effect.
//
// The third bullet is the duration-copy primitive, BecomeCopy (#1593,
// ADR 0043's 2026-09-28 amendment), with no stated duration
// (CopyIndefinite, CR 611.2a). The rulings ask for nothing more: the
// Hallcreeper copies the target's copiable values and nothing else
// (CR 707.2), it is not entering the battlefield so no "enters"
// ability fires, and the copy lasts indefinitely. It keeps its
// counters, and — being the same object — its memory, though as a copy
// it no longer has this trigger to consult it.
//
// "Another" is object identity (effects.Another, CR 109.1): a Clone of this
// creature, or any second creature with the same name, is a legal target.
func init() {
	Register(Spec{
		OracleID:     "6aea681e-88d2-48df-a717-3ce0bc95205b",
		Name:         "Silent Hallcreeper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			silentHallcreeperTrigger(),
		},
	})
}

func silentHallcreeperTrigger() game.TriggeredAbility {
	t := On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
	}, silentHallcreeperLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Put two +1/+1 counters on this creature.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("This creature becomes a copy of another target creature you control.",
			Another(TargetCreature("another target creature you control", YouControl())),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return BecomeCopy{
					Targets:  []uuid.UUID{ctx.Source()},
					Of:       t.ID,
					Duration: CopyIndefinite,
					Label:    "Silent Hallcreeper — becomes a copy",
				}.Apply(ctx)
			}),
	)
	return t
}
