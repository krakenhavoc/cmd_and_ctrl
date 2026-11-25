package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treacherous Blessing — Enchantment {2}{B} (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"When this enchantment enters, draw three cards.
//	 Whenever you cast a spell, you lose 1 life.
//	 When this enchantment becomes the target of a spell or ability, sacrifice
//	 it."
//
// Three separate triggers. The draw is an entry trigger; the life loss is one
// trigger per spell its controller casts (any spell, a copy is not cast);
// and the sacrifice is Phantasmal Image's self-targeted trigger (it resolves
// first, being put on the stack above the targeting spell or ability, and
// does nothing if the enchantment has already left). It targets nothing, so
// the effect that targeted it still resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17f573c4-3ed4-453e-bfbe-4e42d54633e5",
		Name:         "Treacherous Blessing",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Treacherous Blessing — draw three cards", effects.Do(effects.DrawCards{N: 3})),
			effects.WheneverYouCast(nil, "Treacherous Blessing — you lose 1 life", effects.Do(effects.YouLoseLife(1))),
			effects.On(game.EventBecomesTarget, effects.Self, "Treacherous Blessing — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
