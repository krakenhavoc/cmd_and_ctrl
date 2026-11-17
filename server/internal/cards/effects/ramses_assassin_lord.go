package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramses, Assassin Lord — Legendary Creature — Human Assassin {2}{U}{B},
// 4/4:
//
//	"Deathtouch
//	 Other Assassins you control get +1/+1.
//	 Whenever a player loses the game, if they were attacked this turn by an
//	 Assassin you controlled, you win the game."
//
// The anthem is the tribal lord. The trigger watches the elimination event
// (EventPlayerEliminated is emitted before the loser's objects go, so
// Ramses is still there to see it). The intervening "if" (CR 603.4) is
// read as the ability triggers and again as it resolves: the loser was the
// DEFENDER of an attack declared this turn by a creature that was an
// Assassin and that the ability's controller controlled as it was declared
// (Game.AttackedByCreatureTypeThisTurn reads that off the declaration, so
// an Assassin that died in the combat still counts, and an attack on a
// planeswalker or a battle does not). Ramses' own controller losing the
// game cannot trigger it for them: they have left.
//
// No simplification.
func init() {
	trig := InterveningIf(dmuB19LoserAttackedByYourAssassin{},
		On(game.EventPlayerEliminated, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return g.AttackedByCreatureTypeThisTurn(ev.Actor, source.Controller, "Assassin")
		}, "Ramses, Assassin Lord — you win the game", func(g *game.Game, item *game.StackItem) error {
			return WinTheGame{}.Apply(NewContext(g, item))
		}))
	Register(Spec{
		OracleID:        "170182ed-af1d-442a-8868-d74a96e17f03",
		Name:            "Ramses, Assassin Lord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Assassin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{trig},
	})
}

// dmuB19LoserAttackedByYourAssassin is Ramses' intervening "if they were
// attacked this turn by an Assassin you controlled", read for the player
// the triggering event names.
type dmuB19LoserAttackedByYourAssassin struct{}

func (dmuB19LoserAttackedByYourAssassin) Holds(ctx *Context) bool {
	return ctx.Game.AttackedByCreatureTypeThisTurn(ctx.Trigger().Event.Actor, ctx.Controller(), "Assassin")
}

func (dmuB19LoserAttackedByYourAssassin) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
