package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Humiliator — Creature — Vedalken Wizard {3}{U}, 3/4:
//
//	"Metalcraft — Whenever this creature attacks, if you control three or
//	 more artifacts, creatures your opponents control lose all abilities and
//	 have base power and toughness 1/1 until end of turn."
//
// An intervening "if" (CR 603.4): checked as the creature attacks and again
// as the ability resolves. The creatures are locked as it resolves
// (CR 611.2c): ones that enter later keep their abilities and sizes. The
// base 1/1 is layer 7b, so counters and pumps still apply on top.
//
// No simplification.
var vedalkenHumiliatorMetalcraft = effects.CondBoard{Name: "you control three or more artifacts", Test: func(ctx *effects.Context) bool {
	return effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() }) >= 3
}}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "17c3a888-84d2-4c21-8fa3-d2d825dd7601",
		Name:         "Vedalken Humiliator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(vedalkenHumiliatorMetalcraft, effects.On(game.EventAttack,
				effects.AllOf(effects.ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CountControlled(g, source.Controller, func(c game.Card) bool { return c.IsArtifact() }) >= 3
				}),
				"Vedalken Humiliator — creatures your opponents control lose all abilities and have base power and toughness 1/1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mods := append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(1, 1)...)
					return effects.ScopedEffectFor{
						Match:    effects.And(effects.Creature(), effects.OpponentControls()),
						Mods:     mods,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Vedalken Humiliator — lose all abilities and have base power and toughness 1/1",
					}.Apply(ctx)
				})),
		},
	})
}
