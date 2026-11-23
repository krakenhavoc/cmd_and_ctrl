package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenured Inkcaster — Creature — Vampire Warlock {4}{B}, 2/2 (Outlaws of
// Thunder Junction Commander, #5368):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Whenever a creature you control with a +1/+1 counter on it attacks,
//	 each opponent loses 1 life and you gain 1 life."
//
// The attack trigger is one per attacking creature (CR 508.3a); the
// counter is read as the attack is declared, so a counter put on in
// response to the trigger does not retroactively add one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "780536da-2923-48e7-916a-61c701382d28",
		Name:         "Tenured Inkcaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Tenured Inkcaster — put a +1/+1 counter on target creature", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.Cmdr2B23HasPlusOneCounter(g, ev.CardID)
			}, "Tenured Inkcaster — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
