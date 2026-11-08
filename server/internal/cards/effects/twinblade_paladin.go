package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twinblade Paladin — Creature — Human Knight {3}{W}, 3/3 (Foundations,
// #4465):
//
//	"Whenever you gain life, put a +1/+1 counter on this creature.
//	 As long as you have 25 or more life, this creature has double
//	 strike."
//
// The gain trigger is Ajani's Pridemate's, once per life-gain event. The
// double strike is Serra Ascendant's layer-6 grant gated on your life
// total, read live (LifeGatedKeyword), so at a Commander table's 40
// starting life the Paladin has double strike from the start.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d373dc71-5196-43a0-8cc5-67e2a591aad0",
		Name:         "Twinblade Paladin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Twinblade Paladin — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{
			LifeGatedKeyword(SelfWhileYourLifeAtLeast(25), "double strike"),
		},
	})
}
