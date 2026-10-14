package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Delney, Streetwise Lookout — Legendary Creature — Human Scout
// {2}{W}, 2/2:
//
//	"Creatures you control with power 2 or less can't be blocked by
//	 creatures with power 3 or greater.
//	 If a triggered ability of a creature you control with power 2 or
//	 less triggers, that ability triggers an additional time."
//
// Two existing pieces of vocabulary. The first line is Kithkin Armor's
// block rule with the attacker side widened from "enchanted creature"
// to every creature its controller controls with power 2 or less
// (OnMatching). Both powers are read live when blockers are declared
// (CR 509.1b), so a pump after blocks is too late and a pump before
// them takes the creature out of the rule.
//
// The second line is #752's trigger doubler in Harmonic Prodigy's
// shape, DoublesAbilitiesOf with a filter: the source must be a
// creature with power 2 or less, read from its characteristics when
// the ability triggered, so a dies trigger is judged by the power the
// creature last had on the battlefield (CR 603.10). The default
// controller check keeps it to creatures you control. Delney is a 2/2
// and doubles its own triggers too, but it has none.
//
// No simplification.
func init() {
	doubler := DoublesAbilitiesOf(And(Creature(), PowerLE(2)), DoublesAbilitiesOfOptions{})
	doubler.Label = "Delney, Streetwise Lookout"
	doubler.RowLabel = "Triggered abilities of your creatures with power 2 or less trigger an additional time."
	Register(Spec{
		OracleID:     "245d0ccf-87b6-460a-8b99-9e2079f2d375",
		Name:         "Delney, Streetwise Lookout",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnMatching(And(Creature(), YouControl(), PowerLE(2))), PowerGE(3), "creatures with power 3 or greater"),
		},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
