package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurl into History — Instant {3}{U}{U}:
//
//	"Counter target artifact or creature spell. Discover X, where X is
//	 that spell's mana value."
//
// X is read off the spell before it is countered, with the X it was
// cast for counted while it is on the stack (CR 202.3e) — by the time
// the discover runs, it is in a graveyard and that X is gone. The
// discover happens even when the counter does not (a spell that can't
// be countered), because it is a separate sentence. A target that is no
// longer legal fizzles the whole spell (CR 608.2b), so there is nothing
// to discover then. Discover is ADR 0099's (game/discover.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "016718df-2eb5-499f-adea-18231c0375b3",
		Name:         "Hurl into History",
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Targets:      effects.TargetSpell("target artifact or creature spell", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				x := 0
				if spell, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					x, _ = ctx.Game.ManaValueForEffect(spell)
				}
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.Discover{N: x}.Apply(ctx)
			}
			return nil
		},
	})
}
