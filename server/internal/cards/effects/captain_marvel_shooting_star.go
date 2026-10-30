package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Marvel, Shooting Star — Legendary Creature — Human Kree Hero
// {5}{W}{W}, 6/6 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Captain Marvel enters or attacks, exile up to one target
//	 creature. That creature's controller gains life equal to its power.
//	 Whenever a creature other than Captain Marvel is exiled from the
//	 battlefield, you gain life equal to its power."
//
// The first ability is Solitude's body with Swords to Plowshares'
// payout: the controller and the power are read before the exile (CR
// 608.2h). The second watches the battlefield exit itself, so it also
// sees the exile the first ability just made (the opponent gains the
// life, and so does Captain Marvel's controller), a blink, or any other
// exile of a creature; a token is a creature exiled from the battlefield
// too. The power is the creature's last-known power at resolution
// (GainLifeEqualToSubject).
//
// No simplification.
func init() {
	exile := WhenThisEntersOrAttacks(
		"Captain Marvel — exile up to one target creature; its controller gains life equal to its power",
		msc10ExileThenItsControllerGainsItsPower)
	exile.Targets = TargetCreature("up to one target creature").WithCount(0, 1)
	Register(Spec{
		OracleID:        "73fdb375-9c71-421d-8027-8636e3ea8f35",
		Name:            "Captain Marvel, Shooting Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			exile,
			On(game.EventLTB, msc10CreatureExiledFromTheBattlefield,
				"Captain Marvel — you gain life equal to its power",
				Do(GainLifeEqualToSubject{Stat: SubjectPower})),
		},
	})
}
