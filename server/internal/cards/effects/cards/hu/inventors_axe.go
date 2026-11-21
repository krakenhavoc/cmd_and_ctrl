package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inventor's Axe — Artifact — Equipment {R}:
//
//	"Flash
//	 When this Equipment enters, you get {E}{E} (two energy counters).
//	 When this Equipment enters, attach it to target creature you
//	 control.
//	 Equipped creature gets +2/+0.
//	 Equip—Pay {E}{E}."
//
// Two separate enters triggers, as printed: the controller orders them
// (CR 603.3b). The attach is Embercleave's (Targeting over
// WhenThisEnters, resolving through AttachSourceToTarget). Equip is an
// activated ability (CR 702.6a), so "Pay {E}{E}" is an ordinary energy
// component (ADR 0129 §5), checked before anything is paid and never
// waived.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "364fc1a7-b91d-470b-80fe-275828f36cfc",
		Name:            "Inventor's Axe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Energy: 2},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Inventor's Axe", 2),
			effects.Targeting(
				effects.WhenThisEnters("Inventor's Axe — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipPayingAbility("Equip—Pay {E}{E}", effects.PayEnergy(2))},
	})
}
