package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mithril Coat — Legendary Artifact — Equipment for {3}:
//
//	"Flash
//	 Indestructible
//	 When Mithril Coat enters, attach it to target legendary
//	 creature you control.
//	 Equipped creature has indestructible.
//	 Equip {3}"
//
// Removal protection you hold up: flash means it comes down in
// response to the Wrath of God, and the ETB attach means it lands
// already on the commander with no equip activation to pay for and
// no sorcery-speed window to wait for.
//
// Every clause is ordinary machinery, which is the interesting part
// — the card reads like it needs something new and needs nothing:
//
//   - Flash and the Coat's own indestructible are printed keywords.
//     Flash is read off the card in HAND, which is what opens the
//     instant-speed cast; indestructible is read on the battlefield
//     by the destruction path.
//   - The ETB attach is a targeted trigger. The clause rides
//     game.TriggeredAbility.Targets, so the engine picks the target
//     as the ability goes on the stack (CR 603.3d), drops the
//     trigger entirely when there is no legal one, and re-checks the
//     choice at resolution (CR 608.2b). The resolution is
//     AttachSourceToTarget — the same function the equip ability
//     resolves with, because it is the same sentence.
//   - The grant is one GrantToAttached.
//
// "Legendary creature you control" is checked when the trigger picks
// its target. A creature that stops being legendary afterwards keeps
// the Coat, the same way an equipped creature that changes control
// does (CR 702.6c and the note on EquipAbility).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dc364f3-3094-4660-80ca-9418588c7fde",
		Name:            "Mithril Coat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "indestructible"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mithril Coat — attach it to target legendary creature you control",
					AttachSourceToTarget),
				TargetCreature("target legendary creature you control", YouControl(), Legendary()),
			),
		},
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
