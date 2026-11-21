package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forebear's Blade — Artifact — Equipment {3} (Commander Masters):
//
//	"Equipped creature gets +3/+0 and has vigilance and trample.
//	 Whenever equipped creature dies, attach this Equipment to target
//	 creature you control.
//	 Equip {3}"
//
// The pump and keywords are the usual attached statics. The death trigger
// is the Equipment's own (equippedCreatureDied reads the creature that was
// attached as it died, CR 603.10) and targets a creature you control as it
// goes on the stack; attaching is Thunder Lasso's AttachSourceToTarget,
// which does nothing if the Equipment has left or the target is illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9af02f3-6ab1-469c-9d37-e8f6fa9608c4",
		Name:         "Forebear's Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0), effects.GrantToAttached("vigilance", "trample")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.EquippedCreatureDied(ev, source)
				}, "Forebear's Blade — attach this Equipment to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
