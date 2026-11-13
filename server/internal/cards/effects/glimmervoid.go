package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmervoid — Land (Mirrodin, #5606):
//
//	"At the beginning of the end step, if you control no artifacts,
//	 sacrifice this land.
//	 {T}: Add one mana of any color."
//
// The trigger watches every end step, not only its controller's. Its
// "if" is an intervening if (CR 603.4): it triggers only when its
// controller controls no artifacts as the end step begins, and does
// nothing if an artifact has arrived by the time it resolves. A land
// that has already left is not sacrificed. The mana ability is the
// any-colour rock's (samiAnyColorMana).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b92e9854-4527-4133-8615-e282a213e7e3",
		Name:         "Glimmervoid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control no artifacts",
				Test: func(ctx *Context) bool { return !controlsAnyCard(ctx.Game, ctx.Controller(), Artifact()) },
			}, On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b15EndStepBegan(ev) && !controlsAnyCard(g, source.Controller, Artifact())
				},
				"Glimmervoid — if you control no artifacts, sacrifice it",
				SacrificeThisIfStillOnBattlefield)),
		},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}
