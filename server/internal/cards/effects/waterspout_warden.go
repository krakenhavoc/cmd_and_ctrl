package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterspout Warden — Creature — Frog Soldier {2}{U}, 3/2 (Bloomburrow,
// #4467):
//
//	"Whenever this creature attacks, if another creature entered the
//	 battlefield under your control this turn, this creature gains
//	 flying until end of turn."
//
// An intervening "if" (CR 603.4): checked as the attack is declared and
// again as the trigger resolves. "Another creature entered under your
// control this turn" is Bristlebane Outrider's reading of the turn's
// entry tally: a creature entry under your control other than the Warden
// itself, so a Warden that entered this turn needs a second one. A
// Warden that has left the battlefield by then gains nothing, so the
// resolution check reads it only while it is still here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e11386f-4667-47f7-b280-eabd927e4a46",
		Name:         "Waterspout Warden",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "another creature entered the battlefield under your control this turn",
				Test: func(ctx *Context) bool {
					c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && onBattlefield(ctx.Game, c.InstanceID) && bristlebaneOutriderAnotherCreatureEntered(ctx.Game, &c)
				},
			}, On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && bristlebaneOutriderAnotherCreatureEntered(g, source)
			}, "Waterspout Warden — this creature gains flying until end of turn",
				thisCreatureUntilEOT("Waterspout Warden — flying until end of turn", 0, 0, "flying"))),
		},
	})
}
