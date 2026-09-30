package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// choice (Gala Greeters' TriggeredAbility.Modes / ChooseOne, #764) so
// the controller answers a mode_pick prompt rather than the engine
// picking for them.
//
// Two declared simplifications:
//
//   - "That hasn't been chosen" is a per-permanent tally with no seam
//     (ADR 0065 "Out of scope", the same gap Gala Greeters ships
//     with): every bullet is offered every time, so the same mode can
//     be taken on repeated triggers.
//   - The third bullet — "this creature becomes a copy of another
//     target creature you control" — is not offered at all. The
//     catalog's copy machinery (EntersAsCopyOf, S16.5) only replaces
//     an object's OWN entry to the battlefield; there is no primitive
//     for an object already on the battlefield turning into a copy of
//     something else later (that is CR 707's layer-1 "becomes a
//     copy" effect, distinct from "enters as a copy", and unbuilt).
//     Leaving it out is the WEAKER direction the #259 rule requires.
func init() {
	Register(Spec{
		OracleID:     "6aea681e-88d2-48df-a717-3ce0bc95205b",
		Name:         "Silent Hallcreeper",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You choose the mode each time it deals combat damage, but \"that hasn't been chosen\" isn't enforced — the same mode can be chosen more than once.",
			"The \"becomes a copy of another target creature you control\" mode isn't offered — the engine has no way for a permanent already on the battlefield to become a copy of another one later.",
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			silentHallcreeperTrigger(),
		},
	})
}

func silentHallcreeperTrigger() game.TriggeredAbility {
	t := On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
	}, "Silent Hallcreeper — choose one", func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Put two +1/+1 counters on this creature.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	return t
}
