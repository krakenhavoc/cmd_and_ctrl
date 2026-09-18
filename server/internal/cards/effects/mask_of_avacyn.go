package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Avacyn — Artifact — Equipment {2} (EDHREC rank 4085):
//
//	"Equipped creature gets +1/+2 and has hexproof.
//	 Equip {3}"
//
// The cheap protective Equipment a Voltron deck runs when Swiftfoot
// Boots and Lightning Greaves are already in the deck and it still
// wants a third. The +1/+2 is rent; hexproof is the card.
//
// Both halves are the shared attachment statics — PumpAttached for
// the layer 7c bonus, GrantToAttached for the keyword — and hexproof
// is honoured by the engine's targeting check, so an opponent's
// removal spell cannot name the equipped creature at all. It is
// hexproof, not shroud: the controller's own Auras and pump spells
// still target it, which is what makes the card playable rather than
// a liability.
//
// Equip is the shared CR 702.6 ability: sorcery speed, targets a
// creature YOU control, and moves the Mask off whatever it was on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab66f8a8-eb3d-4c2d-95e9-26a53c66b237",
		Name:         "Mask of Avacyn",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantToAttached("hexproof"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
