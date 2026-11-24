package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khalni Heart Expedition — Enchantment {1}{G}:
//
//		"Landfall — Whenever a land you control enters, you may put a quest
//		 counter on this enchantment.
//		 Remove three quest counters from this enchantment and sacrifice it:
//		 Search your library for up to two basic land cards, put them onto the
//		 battlefield tapped, then shuffle."
//
//	  - The landfall trigger is optional ("you may"): the choice is made as it
//	    resolves, and the counter goes on through AddCounter.
//	  - The activation's cost is the removal of exactly three quest counters
//	    and the sacrifice (CR 118.3): with fewer it cannot be activated, and a
//	    fourth counter stays on a Khalni that is already gone with the ability.
//	    The search is Explosive Vegetation's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be6155de-c5b2-415c-ad83-142f9926462a",
		Name:         "Khalni Heart Expedition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Khalni Heart Expedition — put a quest counter on it",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Put a quest counter on Khalni Heart Expedition?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 2, Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), effects.SacrificeThis()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Khalni Heart Expedition — up to two basic lands",
				}.Apply(ctx)
			})),
		}},
	})
}
