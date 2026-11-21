package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Packsong Pup — Creature — Wolf {1}{G}, 1/1 (Innistrad: Crimson Vow,
// #5495):
//
//	"At the beginning of combat on your turn, if you control another Wolf
//	 or Werewolf, put a +1/+1 counter on this creature.
//	 When this creature dies, you gain life equal to its power."
//
// The combat trigger is an intervening "if" (CR 603.4): checked as the
// step begins (the row's AppliesTo) and again as the ability resolves (the
// Do's If), so a Wolf that left in response turns the counter off. The two
// type words are words of this card's text. The dies trigger reads the
// Pup's last-known power at resolution (CR 608.2h), counters included,
// through GainLifeEqualToSubject (Conclave Mentor's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd796e12-c422-4869-8d59-794600425381",
		Name:         "Packsong Pup",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another Wolf or Werewolf",
				Test: func(ctx *effects.Context) bool {
					return effects.VowB07WolvesAndWerewolves(ctx.Game, ctx.Controller(), ctx.TextSubtype("Wolf"), ctx.TextSubtype("Werewolf"), ctx.Source()) > 0
				},
			}, effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) && effects.VowB07AnotherWolfOrWerewolf(g, source)
			}, "Packsong Pup — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
			effects.WhenThisDies("Packsong Pup — you gain life equal to its power",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectPower})),
		},
	})
}
