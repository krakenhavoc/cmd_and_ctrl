package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidus, Blitzball Star — Legendary Creature — Human Warrior {1}{W}{U},
// 2/1 (Final Fantasy, #4462):
//
//	"Whenever an artifact you control enters, put a +1/+1 counter on
//	 Tidus.
//	 Whenever Tidus attacks, tap target creature an opponent controls."
//
// Two triggers on the stack. The first watches each artifact entering
// under Tidus's controller (one trigger per artifact, so a batch of
// Treasures is one counter each) and puts the counter on Tidus only
// while he is still the same object (CounterOnThis). The second is
// Web-Shooters' attack trigger printed on the creature itself: the
// target is chosen as it goes on the stack (CR 603.3d), and with no
// creature an opponent controls it is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef827c62-a5c5-4a7a-92a8-edefb081fae8",
		Name:         "Tidus, Blitzball Star",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Tidus, Blitzball Star — put a +1/+1 counter on Tidus",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			Targeting(
				WheneverThisAttacks("Tidus, Blitzball Star — tap target creature an opponent controls",
					tapTheFirstLegalTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
