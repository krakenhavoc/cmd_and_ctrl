package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deny the Divine — Instant {2}{U} (Theros Beyond Death, #5514):
//
//	"Counter target creature or enchantment spell. If that spell is
//	 countered this way, exile it instead of putting it into its owner's
//	 graveyard."
//
// Dissipate's resolution over a narrower target: a spell that is both
// (an enchantment creature) qualifies, and a spell that can't be countered
// stays on the stack and is not exiled either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0146de73-29bf-415a-b450-11074553f715",
		Name:         "Deny the Divine",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature or enchantment spell", effects.Or(effects.Creature(), effects.Enchantment())),
		OnResolve:    effects.CounterTargetSpellAndExileIt,
	})
}
