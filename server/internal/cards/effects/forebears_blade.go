package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f9af02f3-6ab1-469c-9d37-e8f6fa9608c4",
		Name:         "Forebear's Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 0), GrantToAttached("vigilance", "trample")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return equippedCreatureDied(ev, source)
				}, "Forebear's Blade — attach this Equipment to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
