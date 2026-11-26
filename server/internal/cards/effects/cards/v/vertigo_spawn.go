package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vertigo Spawn — Creature — Illusion {1}{U}, 0/3 (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Defender
//	 Whenever this creature blocks a creature, tap that creature. That
//	 creature doesn't untap during its controller's next untap step."
//
// Cleric of Chill Depths's trigger (WheneverThisBlocksItDoesntUntap): once
// for each attacker it blocks (CR 509.3b), and "that creature" is the
// attacker as it was when blocked, so one that left and came back before the
// trigger resolved is a new object and is untouched (CR 400.7). The tap comes
// first, then the untap lock.
//
// No simplification.
func init() {
	const label = "Vertigo Spawn — tap the creature it blocks; that creature doesn't untap during its controller's next untap step"
	t := effects.WheneverThisBlocksItDoesntUntap(label)
	lock := t.Effect
	t.Effect = func(g *game.Game, item *game.StackItem) error {
		attacker := item.Params.Object
		if attacker.ID != uuid.Nil && effects.OnBattlefield(g, attacker.ID) && g.ObjectEpochForEffect(attacker.ID) == attacker.Epoch {
			if err := (effects.TapTarget{Target: attacker.ID}).Apply(effects.NewContext(g, item)); err != nil {
				return err
			}
		}
		return lock(g, item)
	}
	effects.Register(effects.Spec{
		OracleID:        "c1b8cfe9-e8fc-4321-94ce-86ab1f1ae2d0",
		Name:            "Vertigo Spawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
