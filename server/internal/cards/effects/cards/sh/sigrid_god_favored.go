package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigrid, God-Favored — Legendary Creature — Human Warrior {1}{W}{W}, 2/2:
//
//	"Flash
//	 First strike, protection from God creatures
//	 When Sigrid enters, exile up to one target attacking or blocking
//	 creature until Sigrid leaves the battlefield."
//
// Banisher Priest's exile until this leaves, over up to one target attacking
// or blocking creature of any controller; with no target nothing is exiled.
// "Protection from God creatures" is the creature-type quality (a God that is
// not a creature does not stop it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84d37104-2681-42c1-9d19-91a2ebe6a435",
		Name:            "Sigrid, God-Favored",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "first strike", "protection from God creatures"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sigrid, God-Favored — exile up to one target attacking or blocking creature until Sigrid leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Sigrid, God-Favored — the exiled creature returns when Sigrid leaves the battlefield")),
				effects.UpToOneTargetCreature("up to one target attacking or blocking creature", effects.AttackingOrBlocking())),
		},
	})
}
