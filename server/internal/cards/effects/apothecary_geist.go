package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apothecary Geist — Creature — Spirit {3}{W}, 2/3 (Innistrad Remastered):
//
//	"Flying
//	 When this creature enters, if you control another Spirit, you gain 3 life."
//
// An intervening "if" (CR 603.4), written as Dwynen's Elite writes his: the
// other Spirit is looked for as the creature enters and again as the
// ability resolves. "Another" excludes this creature by instance.
func init() {
	Register(Spec{
		OracleID:        "3420da71-0af4-46b4-a994-91a82332d0fb",
		Name:            "Apothecary Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another Spirit",
				Test: func(ctx *Context) bool {
					return mshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Spirit"))
				},
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && mshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Spirit"))
			}, "Apothecary Geist — you gain 3 life", Do(GainLife{Amount: 3}))),
		},
	})
}
