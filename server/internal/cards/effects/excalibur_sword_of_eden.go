package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excalibur, Sword of Eden — Legendary Artifact — Equipment for
// {12}:
//
//	"This spell costs {X} less to cast, where X is the total mana
//	 value of historic permanents you control. (Artifacts,
//	 legendaries, and Sagas are historic.)
//	 Equipped creature gets +10/+0 and has vigilance.
//	 Equip legendary creature {2}"
//
// A twelve-mana card that a built board casts for four or five, and
// then a two-mana equip turns a commander into a lethal attacker.
//
// The cost clause is a SELF cost modifier (Spec.SelfCostModifiers,
// ADR 0048 addendum) and not a caveat, because the slot already
// exists and takes an amount computed per cast — the same shape The
// Great Henge and Ghalta use. Three properties come with it for
// free and all three are printed:
//
//   - It is a REDUCTION, so it spends against generic mana only and
//     floors at zero (CR 601.2f). Excalibur's cost is all generic,
//     so a big enough board really does make it free.
//   - It counts TOTAL MANA VALUE, not permanents. One Sol Ring is
//     one; one eight-drop legend is eight.
//   - It never counts Excalibur itself. The engine prices the spell
//     while the card is still in its source zone, and the counting
//     helper skips the card being cast, so a second copy already on
//     the battlefield counts and this one does not.
//
// "Historic" is CR 700.6 — an artifact, a legendary, or a Saga —
// read post-layer, so an animated artifact land counts and a token
// contributes its (absent) mana value of zero.
//
// The equip ability is legendary-only, the tighter target clause
// Blackblade Reforged also carries. It is checked WHILE ACTIVATING
// (CR 702.6c): a creature that stops being legendary afterwards
// keeps the sword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fb6bd72-031b-40a1-83c5-8a1c82f84e12",
		Name:         "Excalibur, Sword of Eden",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(TotalManaValueOfHistoricPermanentsYouControl(),
				"This spell costs {X} less to cast, where X is the total mana value of historic permanents you control."),
		},
		Static: []game.StaticAbility{
			PumpAttached(10, 0),
			GrantToAttached("vigilance"),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip legendary creature {2}", "{2}",
				TargetCreature("target legendary creature you control", YouControl(), Legendary())),
		},
	})
}
