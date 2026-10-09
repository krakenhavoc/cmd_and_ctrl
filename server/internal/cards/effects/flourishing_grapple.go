package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flourishing Grapple — Instant {G}:
//
//	"Target creature or planeswalker an opponent controls that's red or
//	 white loses all abilities until end of turn. Target creature you
//	 control deals damage equal to its power to that permanent."
//
// Bite Down with a rider, read off a two-clause announcement: slot 0 is
// the red-or-white permanent, slot 1 the creature that bites. The
// ability loss is a layer-6 scoped effect on slot 0 and lasts the turn,
// so it is already in force when the damage is dealt (a planeswalker
// has no abilities left to use afterwards, a creature loses its
// keywords). Each half needs its own target: with the opponent's
// permanent gone nothing happens, and with the biter gone the
// permanent still loses its abilities but takes no damage (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6b4f1569-2025-46a8-89ba-79859208b04e",
		Name:         "Flourishing Grapple",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target creature or planeswalker an opponent controls that's red or white",
				And(Or(Creature(), Planeswalker()), OpponentControls(), Or(OfColor("R"), OfColor("W")))),
			TargetCreature("target creature you control", YouControl()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := ctx.ClauseTarget(0)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			if err := untilEndOfTurn(ctx, victim.ID, nil,
				"Flourishing Grapple — loses all abilities", game.LoseAllAbilitiesMod()); err != nil {
				return err
			}
			biter, ok := ctx.ClauseTarget(1)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}
