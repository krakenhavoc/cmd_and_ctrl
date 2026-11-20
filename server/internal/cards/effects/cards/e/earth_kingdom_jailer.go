package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earth Kingdom Jailer — Creature — Human Soldier Ally {2}{W}, 3/3
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature enters, exile up to one target artifact,
//	 creature, or enchantment an opponent controls with mana value 3 or
//	 greater until this creature leaves the battlefield."
//
// Banisher Priest's exile (exileChosenTargetUntilThisLeaves, CR 610.3)
// with a wider, "up to one" clause. If the Jailer has left before the
// trigger resolves, the target is not exiled at all (the ruling; CR
// 610.3c, ExileUntil's ThisLeaves). The exiled card returns as a new
// object (CR 400.7) when the Jailer leaves; an Aura chooses what it
// enchants as it returns.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b799c30-4a43-403d-b822-0555bd513913",
		Name:         "Earth Kingdom Jailer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Earth Kingdom Jailer — exile up to one target artifact, creature, or enchantment an opponent controls with mana value 3 or greater until this creature leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Earth Kingdom Jailer — the exiled card returns when Earth Kingdom Jailer leaves the battlefield")),
				effects.TargetPermanent("up to one target artifact, creature, or enchantment an opponent controls with mana value 3 or greater",
					effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()), effects.OpponentControls(), effects.ManaValueGE(3)).WithCount(0, 1)),
		},
	})
}
