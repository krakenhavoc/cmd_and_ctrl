package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodthorn Flail — Artifact — Equipment {B} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Equipped creature gets +2/+1.
//	 Equip—Pay {3} or discard a card."
//
// Equip is an activated ability (CR 702.6a) whose cost is one either/or
// part (game.AbilityCost.Either): the activator announces which branch they
// pay, {3} from the pool or one card discarded from their hand, with the
// rest of the announcement (CR 602.2b, 601.2b), and a branch they can't pay
// can't be chosen (CR 118.3). The mana branch (game.AbilityCostBranchMana)
// folds into the ability's mana cost, so the pool payment is the ordinary
// one. Equip is sorcery speed as ever.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01ab7724-1967-4da6-aebd-3729983d6a0e",
		Name:         "Bloodthorn Flail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Activated: []ActivatedAbility{
			EquipPayingAbility("Equip—Pay {3} or discard a card.",
				OrCost(ManaBranch("{3}"), DiscardBranch(1, "a card", false))),
		},
	})
}
