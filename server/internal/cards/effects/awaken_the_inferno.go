package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awaken the Inferno — Sorcery {4}{R}:
//
//	"Awaken the Inferno deals 6 damage to target creature or planeswalker
//	 an opponent controls. Put a +1/+1 counter on up to one target
//	 creature you control.
//	 Basic landcycling {2} ({2}, Discard this card: Search your library
//	 for a basic land card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Two clauses: slot 0 is the burn target, slot 1 is "up to one" creature
// you control (Min 0, so the spell can be cast with none). The halves are
// independent (CR 608.2b): the counter lands even when the burn target
// has gone, and the burn lands without a counter target. Basic
// landcycling is the shared hand ability (Sylvan Reclamation).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e729fce5-a1cf-4e95-9fd5-fccf4f704d31",
		Name:         "Awaken the Inferno",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target creature or planeswalker an opponent controls",
				Or(Creature(), Planeswalker()), OpponentControls()),
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		),
		Purpose: ForTargets(DamageToTarget(0, 6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 6}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
			return nil
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
