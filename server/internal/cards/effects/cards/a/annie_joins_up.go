package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	doubler := effects.DoublesAbilitiesOf(effects.And(effects.Creature(), effects.Legendary()), effects.DoublesAbilitiesOfOptions{})
	doubler.Label = "Annie Joins Up"
	doubler.RowLabel = "Triggered abilities of your legendary creatures trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:     "4c3ad545-b375-44a7-a24e-58baacc4e4b6",
		Name:         "Annie Joins Up",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Annie Joins Up — 5 damage to target creature or planeswalker an opponent controls", effects.SourceDealsDamageToEachLegalTarget(5)),
				effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 5))),
		},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
