package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbarian Class — Enchantment — Class {R} (Adventures in the Forgotten
// Realms, #5681):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 If you would roll one or more dice, instead roll that many dice plus
//	 one and ignore the lowest roll.
//	 {1}{R}: Level 2
//	 Whenever you roll one or more dice, target creature you control gets
//	 +2/+0 and gains menace until end of turn.
//	 {2}{R}: Level 3
//	 Creatures you control have haste."
//
// One line per level (ADR 0071):
//
//   - Level 1 is Pixie Guide's replacement on the dice roll's window
//     (RollOneMoreDieAndIgnoreTheLowest), ungated: a Class has its top
//     ability at every level (CR 716.3).
//   - Level 2 triggers once per instruction to roll, however many dice it
//     rolled (WheneverYouRollDice), and only for a roll that kept a die:
//     an ignored roll never happened (CR 706.6). It targets as it is put
//     on the stack, so with no creature it is removed from the stack (CR
//     603.3d). A planar die is a die rolled (CR 706.7).
//   - Level 3 is Fervor's layer-6 haste grant, gated at level 3.
//
// No simplification.
func init() {
	haste := Labelled(TribalKeywordGrant(TribeFilter{YoursOnly: true}, "haste"), "Creatures you control have haste.")
	haste.ActiveWhen = Level(3)
	Register(Spec{
		OracleID:     "59faa1b9-17aa-4f2c-a8f8-17ab50392b36",
		Name:         "Barbarian Class",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{RollOneMoreDieAndIgnoreTheLowest("Barbarian Class")},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{R}")),
			LevelUp(3, ManaCost("{2}{R}")),
		},
		Triggered: []game.TriggeredAbility{
			AtLevel(2, TriggerWithPurpose(Targeting(
				WheneverYouRollDice("Barbarian Class — target creature you control gets +2/+0 and gains menace until end of turn",
					func(g *game.Game, item *game.StackItem, _ int) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Barbarian Class — +2/+0"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"menace"}, Label: "Barbarian Class — menace"}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 2, 0, "menace")))),
		},
		Static: []game.StaticAbility{haste},
	})
}
