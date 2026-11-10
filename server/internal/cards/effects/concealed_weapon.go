package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Concealed Weapon — Artifact — Equipment {1}{R} (Murders at Karlov
// Manor, #4469):
//
//	"Equipped creature gets +3/+0.
//	 Disguise {2}{R} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this Equipment is turned face up, attach it to target creature
//	 you control.
//	 Equip {1}{R}"
//
// An Equipment with disguise: the face-down cast is the shared keyword
// (Disguise, a 2/2 creature with ward {2} until it is turned up, CR
// 702.168), and turning it up is Essence of Antiquity's trigger
// (WhenThisIsTurnedFaceUp). The trigger's target is chosen as it goes on
// the stack; it attaches the Equipment to that creature by an effect, so
// no equip cost is paid and the sorcery-speed restriction does not apply
// (the ruling). A target that has become illegal leaves the Equipment
// unattached on the battlefield (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "104565b1-2db8-4c01-ae3d-14087768be45",
		Name:             "Concealed Weapon",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{2}{R}")},
		Static:           []game.StaticAbility{PumpAttached(3, 0)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Concealed Weapon — attach it to target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						host := FirstLegalBattlefieldTarget(NewContext(g, item))
						if host == uuid.Nil || !onBattlefield(g, item.SourceCardID) {
							return nil
						}
						return g.AttachForEffect(item.SourceCardID, game.TargetRef{Kind: game.TargetCard, ID: host})
					}),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}{R}")},
	})
}
