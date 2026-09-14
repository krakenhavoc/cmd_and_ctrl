package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Vengeance — Artifact — Equipment for {3} (EDHREC rank
// 1587):
//
//	"Equipped creature gets +2/+0 and has first strike, vigilance,
//	 trample, and haste.
//	 Equip {3}"
//
// Four keywords on one grant, and every one of them is in the
// engine's honoured table with a live consumer: first strike splits
// the combat damage step, vigilance skips the attack tap, trample
// assigns the excess, haste clears the summoning-sickness gate. The
// card is here precisely because it is the widest single
// GrantToAttached in the batch — if a keyword had been missing from
// the table this is the file it would have shown up in.
//
// Haste is the half that makes it a Voltron card rather than a stat
// stick: it turns a freshly-cast commander into an attacker the turn
// it lands, which is the play the whole archetype is built around.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e366afb3-c447-4bde-b358-41c8568142d5",
		Name:         "Sword of Vengeance",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("first strike", "vigilance", "trample", "haste"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
