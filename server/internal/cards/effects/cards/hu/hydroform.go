package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hydroform — Instant {G}{U} (Gatecrash, S82 OLD #5565):
//
//	"Target land becomes a 3/3 Elemental creature with flying until end of
//	 turn. It's still a land."
//
// Blinkmoth Nexus's animation on a targeted land: Creature added to its types
// and Elemental to its subtypes (layer 4), flying (layer 6) and a 3/3 base
// (layer 7b), all until end of turn, keeping everything it was, colour and mana
// abilities included. The set is the one land, locked at resolution
// (CR 611.2c). A land a creature that came under its controller's control this
// turn can't attack or tap for {T} abilities (CR 302.6) — summoning sickness is
// the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b6ea82-da4e-4763-84b1-4caa214ed945",
		Name:         "Hydroform",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.ScopedEffectFor{
				Target: t.ID,
				Mods: append([]game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
					game.AddKeywordsMod("flying"),
				}, game.SetBasePTMods(3, 3)...),
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Hydroform — a 3/3 Elemental creature with flying until end of turn",
			}.Apply(ctx)
		},
	})
}
