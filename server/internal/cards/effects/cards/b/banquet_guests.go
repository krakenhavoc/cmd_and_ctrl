package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banquet Guests — Creature — Halfling Citizen {X}{G}{W}, 0/0 (Tales of
// Middle-earth Commander):
//
//	"Affinity for Foods (This spell costs {1} less to cast for each Food
//	 you control.)
//	 Trample
//	 This creature enters with twice X +1/+1 counters on it.
//	 {2}, Sacrifice a Food: This creature gains indestructible until end
//	 of turn."
//
// Affinity is the shared constructor over Foods by effective subtype
// (CR 702.41a, reducing only the generic part, so the {X} is paid first
// by the reduction); the entry counters are read off the announced X
// (CR 107.3m, CR 614.1c) as a self-replacement; "Sacrifice a Food" is
// Bosco's cost, any Food you control, a token or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7f2579ef-0422-4cb1-a0f8-6ad0d942515b",
		Name:            "Banquet Guests",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Foods", effects.HasSubtype("Food")),
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.S82GapB117TwiceXCounters(game.CounterPlusOne)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a Food: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Banquet Guests — indestructible",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
