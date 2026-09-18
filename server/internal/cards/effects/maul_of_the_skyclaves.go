package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maul of the Skyclaves — Artifact — Equipment {2}{W} (EDHREC rank
// 4089):
//
//	"When this Equipment enters, attach it to target creature you
//	 control.
//	 Equipped creature gets +2/+2 and has flying and first strike.
//	 Equip {2}{W}{W}"
//
// The Equipment that does not cost you a turn. Every other Equipment
// in this batch asks for the cast and then the equip; the Maul
// attaches itself on the way in, so three mana buys +2/+2, flying
// and first strike on a creature that is already on the board and
// can attack the same turn. The punitive {2}{W}{W} equip cost is the
// price of that, and it is why the Maul is a cast-it-once card: if
// the creature dies you would rather draw a new Maul than move this
// one.
//
// THE ENTRY CLAUSE IS A TARGETED TRIGGER, NOT AN AsEnters HOOK. "When
// this Equipment enters" uses the stack (CR 603.6a), names a target
// on announce, and re-checks it on resolution — so an opponent can
// respond by killing the creature you pointed at and the Maul lands
// unattached, which is exactly what the printed card does. It is
// built from WhenThisEnters wrapped in Targeting, and its body is the
// same AttachSourceToTarget the equip ability resolves with: one
// attach path, one CR 608.2b illegal-target skip.
//
// The target clause is "target creature you control", so the Maul
// cannot be donated to an opponent's creature on the way in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "088a1f56-198c-41a1-b244-ec1c7a771041",
		Name:         "Maul of the Skyclaves",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying", "first strike"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Maul of the Skyclaves — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}{W}{W}"),
		},
	})
}
