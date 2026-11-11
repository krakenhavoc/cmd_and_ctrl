package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "780536da-2923-48e7-916a-61c701382d28",
		Name:         "Tenured Inkcaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Tenured Inkcaster — put a +1/+1 counter on target creature", PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && cmdr2B23HasPlusOneCounter(g, ev.CardID)
			}, "Tenured Inkcaster — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
