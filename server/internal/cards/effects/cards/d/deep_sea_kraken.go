package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep-Sea Kraken — Creature — Kraken {7}{U}{U}{U}, 6/6 (Time Spiral, #5596):
//
//	"This creature can't be blocked.
//	 Suspend 9—{2}{U} (Rather than cast this card from your hand, you
//	 may pay {2}{U} and exile it with nine time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)
//	 Whenever an opponent casts a spell, if this card is suspended,
//	 remove a time counter from it."
//
// An unblockable 6/6 whose suspend clock runs faster the more the table
// casts. The third ability is Nihilith's shape: a trigger that works from
// exile, with the intervening "if this card is suspended" (CR 702.62b, CR
// 603.4) read as it triggers. It is not optional, and it reuses Nihilith's
// body, which removes a counter only from the same exiled object (CR
// 400.7). Removing the last counter this way offers the free cast through
// the same road the upkeep tick uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54de4a7f-86f4-4029-80f7-bc25d6509cc8",
		Name:         "Deep-Sea Kraken",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(9, "{2}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.InExile(effects.On(game.EventCast, effects.TspB1OpponentCastWhileSuspended,
				"Deep-Sea Kraken — remove a time counter from this card", effects.NihilithRemoveTimeCounter)),
		},
	})
}
