package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drake Hatcher — Creature — Human Wizard {1}{U}, 1/3 (Foundations,
// #4465):
//
//	"Vigilance, prowess
//	 Whenever this creature deals combat damage to a player, put that
//	 many incubation counters on it.
//	 Remove three incubation counters from this creature: Create a 2/2
//	 blue Drake creature token with flying."
//
// Vigilance and prowess are engine keywords. "That many" is the damage
// actually dealt, read off the damage event (CR 119.3), as Necropolis
// Regent's is; the counters go on the Hatcher itself. The activation is
// paid by removing the counters as it is announced (CR 602.2b), so it
// cannot be responded to by removing them, and it has no tap symbol, so a
// freshly cast Hatcher may use it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4de0ccab-bb5d-4c7c-827e-36aa1d00b182",
		Name:            "Drake Hatcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Drake Hatcher — put that many incubation counters on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: "incubation", N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove three incubation counters from this creature: Create a 2/2 blue Drake creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    RemoveCountersFromThis("incubation", 3),
			Effect:  Do(CreateToken{Template: TokenCard("2/2 blue Drake with flying"), N: 1}),
		}},
	})
}
