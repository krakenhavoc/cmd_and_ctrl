package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Annie Joins Up — Legendary Enchantment {1}{R}{G}{W} (Marvel Super
// Heroes Commander, #4455):
//
//	"When Annie Joins Up enters, it deals 5 damage to target creature or
//	 planeswalker an opponent controls.
//	 If a triggered ability of a legendary creature you control triggers,
//	 that ability triggers an additional time."
//
// The doubler is Delney's shape (CR 603.2d) with a legendary-creature
// source filter, read off the source's effective type line, so a creature
// made legendary by an effect counts. The enchantment's own ETB is not a
// creature's trigger and is not doubled.
//
// No simplification.
func init() {
	doubler := DoublesAbilitiesOf(And(Creature(), Legendary()), DoublesAbilitiesOfOptions{})
	doubler.Label = "Annie Joins Up"
	doubler.RowLabel = "Triggered abilities of your legendary creatures trigger an additional time."
	Register(Spec{
		OracleID:     "4c3ad545-b375-44a7-a24e-58baacc4e4b6",
		Name:         "Annie Joins Up",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Annie Joins Up — 5 damage to target creature or planeswalker an opponent controls", sourceDealsDamageToEachLegalTarget(5)),
				TargetPermanent("target creature or planeswalker an opponent controls", Or(Creature(), Planeswalker()), OpponentControls())),
				ForTargets(DamageToTarget(0, 5))),
		},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
