package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Marauder — Creature — Ogre Warrior {1}{B}{B}, 3/1 (Betrayers of Kamigawa, S82 OLD
// b595):
//
//	"Whenever this creature attacks, it gains "this creature can't be blocked" until end of
//	 turn unless defending player sacrifices a creature of their choice."
//
// Demanding Dragon's punisher: the defending player decides between letting it become
// unblockable and sacrificing a creature (their choice of which); with no creature to
// sacrifice it is simply unblockable. Blockers are declared after the trigger resolves, so
// the restriction is in place in time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9dafa31-41ea-4f1c-bc63-ec84b01d57fa",
		Name:         "Ogre Marauder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ogre Marauder — it can't be blocked this turn unless defending player sacrifices a creature", effects.OldB595OgreMarauderAsk),
		},
	})
}
