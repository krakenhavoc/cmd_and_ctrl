package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phlage, Titan of Fire's Fury — Legendary Creature — Elder Giant
// {1}{R}{W}, 6/6:
//
//	"When Phlage enters, sacrifice it unless it escaped.
//	 Whenever Phlage enters or attacks, it deals 3 damage to any
//	 target and you gain 3 life.
//	 Escape—{R}{R}{W}{W}, Exile five other cards from your graveyard."
//
// The card #653 was written for, and the cleanest statement of what
// CR 400.7d is: three mana buys you a Lightning Helix and nothing
// else, because the 6/6 sacrifices itself — UNLESS the spell that
// became it was cast for its escape cost. The permanent has to
// remember how it was cast, and until #653 nothing on a permanent
// did.
//
// `ctx.Escaped()` is that memory (CR 702.138b). Note which object it
// asks about: the SOURCE permanent, not the stack item being resolved.
// By the time this trigger resolves the item on the stack is the
// trigger, and the spell that paid escape finished resolving two steps
// ago — which is exactly why ctx.PaidAltCost, the reader an overloaded
// Cyclonic Rift uses, is the wrong one here.
//
// The two triggers are independent (CR 603.3) and both fire on the
// entry. The sacrifice does not cancel the Helix: a hard-cast Phlage
// deals its 3 and gains its 3, then dies, and the order the controller
// puts them on the stack in does not change either outcome.
//
// The sacrifice trigger checks the battlefield first. It sat on the
// stack and anybody could answer it — a Phlage already exiled or
// bounced is not there to sacrifice, and CR 608.2c says the ability
// does as much as it can.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3407eb6e-b74d-4159-a801-d7163937953c",
		Name:             "Phlage, Titan of Fire's Fury",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{R}{R}{W}{W}", 5)},
		Triggered: []game.TriggeredAbility{
			SacrificeThisUnlessItEscaped("Phlage"),
			Targeting(
				WhenThisEntersOrAttacks("Phlage — 3 damage to any target, gain 3 life", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) > 0 {
						if err := (DealDamage{
							Source: item.SourceCardID,
							Target: item.Targets[0].ID,
							Amount: 3,
						}).Apply(ctx); err != nil {
							return err
						}
					}
					// The life is not conditional on the damage — "and
					// you gain 3 life" is a second instruction, so a
					// target that left in response costs the damage
					// and not the life (CR 608.2c).
					return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
				}),
				TargetAny(),
			),
		},
	})
}
