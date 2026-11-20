package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trespassing Souleater — Artifact Creature — Phyrexian Construct {3}, 2/2 (New Phyrexia, #5575):
//
//	"{U/P}: This creature can't be blocked this turn. ({U/P} can be paid with either {U} or 2 life.)"
//
// Drownyard Amalgam's row for {U/P}: a can't-be-blocked restriction on
// the source until end of turn, which does nothing to a source that left
// and came back (CR 400.7). The Phyrexian symbol is paid with {U} or 2
// life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d19b4d1-49d1-44e9-a453-dcb08b5fa6e4",
		Name:         "Trespassing Souleater",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U/P}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{U/P}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Trespassing Souleater — can't be blocked this turn"}.Apply(NewContext(g, item))
			},
		}},
	})
}
