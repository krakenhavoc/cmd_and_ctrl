package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
var vedalkenHumiliatorMetalcraft = CondBoard{Name: "you control three or more artifacts", Test: func(ctx *Context) bool {
	return countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() }) >= 3
}}

func init() {
	Register(Spec{
		OracleID:     "17c3a888-84d2-4c21-8fa3-d2d825dd7601",
		Name:         "Vedalken Humiliator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(vedalkenHumiliatorMetalcraft, On(game.EventAttack,
				AllOf(ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return countControlled(g, source.Controller, func(c game.Card) bool { return c.IsArtifact() }) >= 3
				}),
				"Vedalken Humiliator — creatures your opponents control lose all abilities and have base power and toughness 1/1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					mods := append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(1, 1)...)
					return ScopedEffectFor{
						Match:    And(Creature(), OpponentControls()),
						Mods:     mods,
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Vedalken Humiliator — lose all abilities and have base power and toughness 1/1",
					}.Apply(ctx)
				})),
		},
	})
}
