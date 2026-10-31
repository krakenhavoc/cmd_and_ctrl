package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Human Torch, Johnny Storm — {2}{R} Legendary Creature — Human Hero 2/2
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Whenever you draw a card, if you control another Hero, Human Torch
//	 deals 1 damage to target opponent.
//	 Power-up — {6}{R}: Put three +1/+1 counters on Human Torch. (Activate
//	 each power-up ability only once. Reduce the cost by his mana cost if he
//	 entered this turn.)"
//
// The draw trigger fires once for each card you draw, and its "if" is the
// CR 603.4 intervening kind: read when the card is drawn and again as the
// ability resolves, so losing the other Hero in response fizzles the damage
// (the opponent is still chosen as the ability goes on the stack). Power-up
// is the engine's (CR 702.193): the turn he enters it costs {6}{R} less
// {2}{R}, which is {4}; after that {6}{R}.
//
// No simplification.
func init() {
	powerUp := PowerUp("{6}{R}: Put three +1/+1 counters on Human Torch.", ManaCost("{6}{R}"), plusOneCountersOnThis(3))
	powerUp.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "5b6f2770-e774-4a77-801f-a3fda093c3dd",
		Name:            "Human Torch, Johnny Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(CondBoard{
				Name: "you control another Hero",
				Test: func(ctx *Context) bool {
					return mshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Hero"))
				},
			}, On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && mshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Hero"))
			}, "Human Torch — 1 damage to target opponent", b36DesertDualDamage)),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{powerUp},
	})
}
