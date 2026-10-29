package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avengers Assemble! — Enchantment {4}{W} (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Heroes you control get +2/+2.
//	 At the beginning of each end step, if you attacked with a Hero this
//	 turn or a Hero entered the battlefield under your control this turn,
//	 draw a card."
//
// The anthem is a TribeFilter over the Hero creature type, restricted to
// the controller's creatures (the enchantment itself is no creature, so
// "Heroes" cannot include it). "Each end step" is every player's, so the
// trigger does not read whose end step it is; the intervening "if"
// (CR 603.4) is checked when the trigger would fire and again as it
// resolves, against the turn's tallies: the subtypes a creature had as it
// was declared an attacker (AttackedWithSubtypeThisTurn) or as it entered
// (EnteredWithSubtypeThisTurn, a token's entry counts). Both are the
// controller's own, so a Hero that attacked or entered on an opponent's
// turn still draws on that turn's end step, as printed.
//
// No simplification.
func init() {
	yourHeroes := TribeFilter{Tribes: []string{"Hero"}, YoursOnly: true}
	Register(Spec{
		OracleID:        "3477e1cc-46af-46f4-9ea2-bdf0bf1f880e",
		Name:            "Avengers Assemble!",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{TribalAnthem(yourHeroes, 2, 2)},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && mshA1HeroActedThisTurn(g, source.Controller, source.TextSubtype("Hero"))
			}, "Avengers Assemble! — draw a card", Do(If{Cond: mshA1HeroActedCond, Then: []Applier{DrawCards{N: 1}}})),
		},
	})
}

// mshA1HeroActedCond is Avengers Assemble!'s intervening "if" as it
// resolves (CR 603.4); the row's AppliesTo reads the same fact as it
// triggers. It reads this turn's tallies, which are the turn's event log.
var mshA1HeroActedCond = CondBoard{
	Name: "you attacked with a Hero this turn or a Hero entered under your control this turn",
	Test: func(ctx *Context) bool {
		return mshA1HeroActedThisTurn(ctx.Game, ctx.Controller(), ctx.TextSubtype("Hero"))
	},
}

// mshA1HeroActedThisTurn is Avengers Assemble!'s condition: the player
// attacked with a Hero this turn, or a Hero entered under their control
// this turn.
func mshA1HeroActedThisTurn(g *game.Game, player uuid.UUID, hero string) bool {
	return g.AttackedWithSubtypeThisTurn(player, hero) > 0 || g.EnteredWithSubtypeThisTurn(player, hero) > 0
}
