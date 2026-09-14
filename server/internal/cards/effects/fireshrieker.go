package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fireshrieker — Artifact — Equipment for {3} (EDHREC rank 1120):
//
//	"Equipped creature has double strike.
//	 Equip {2}"
//
// Battle Mastery in artifact form, and the reason to have both is
// that a deck which wants double strike wants it on an Equipment it
// can move after the first carrier dies — an Aura is a two-for-one
// waiting to happen and this is not.
//
// Double strike is honoured: the combat damage step runs twice, and a
// creature carrying both first strike and double strike still gets
// exactly two hits rather than three. It multiplies everything else
// in the attachment catalog — the Warhammer's +3/+0, a Sword's
// combat-damage trigger, Quietus Spike's halving — which is what
// makes it worth its three-and-two mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a02e1ca7-23c5-41e3-a744-72fc9e9dd8ba",
		Name:         "Fireshrieker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("double strike"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
