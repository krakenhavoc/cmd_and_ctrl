package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blackblade Reforged — Legendary Artifact — Equipment for {2}
// (EDHREC rank 349):
//
//	"Equipped creature gets +1/+1 for each land you control.
//	 Equip legendary creature {3}
//	 Equip {7}"
//
// The Voltron finisher: in a deck with ten lands on the board this is
// a +10/+10 for three mana onto a commander, and commander damage
// does the rest.
//
// TWO NEW SHAPES, both of them compositions rather than machinery:
//
//   - THE BONUS IS COUNTED, NOT FIXED. PumpAttachedPer re-evaluates
//     the land count on every layer recompute, so playing a land
//     grows the creature in the same beat and a Strip Mine on your
//     own land shrinks it. Nothing is captured at equip time.
//     "You control" is the EQUIPMENT's controller, which is what CR
//     109.5 says and what makes a stolen carrier still count its real
//     owner's lands.
//   - TWO EQUIP ABILITIES. Equip is an ordinary Spec.Activated entry
//     (ADR 0036 decision 4), so a second one with a tighter target
//     clause and a cheaper cost needs nothing but a second entry. The
//     legendary-only ability is listed FIRST so the cheaper option is
//     the one a player reaches for; both are gated at sorcery speed
//     and both re-check their target at resolution.
//
// The restriction is checked WHILE ACTIVATING, per CR 702.6c and the
// note on EquipAbility: a creature that stops being legendary after
// the cheap equip resolves keeps the Blackblade.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dca51281-fb21-45b6-beb4-1f13397caee2",
		Name:         "Blackblade Reforged",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, landsControlledBy),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip legendary creature {3}", "{3}",
				TargetCreature("target legendary creature you control", YouControl(), Legendary())),
			EquipAbility("{7}"),
		},
	})
}
