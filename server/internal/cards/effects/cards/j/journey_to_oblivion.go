package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Journey to Oblivion — Enchantment {4}{W} (Zendikar Rising, #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Mystical Tether's one-shot exile (CR 610.3): the target is chosen as the
// trigger goes on the stack, and if the enchantment has already left when
// the ability resolves nothing is exiled at all (CR 610.3c). The party is
// counted as the spell is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "31e8c1e1-520a-4e64-a221-203afc819fcf",
		Name:              "Journey to Oblivion",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{effects.OldB095PartyCostsLess()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Journey to Oblivion — exile target nonland permanent an opponent controls until this enchantment leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Journey to Oblivion — the exiled card returns when Journey to Oblivion leaves the battlefield")),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
	})
}
