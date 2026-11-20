package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonglove Changeling — Creature — Shapeshifter {2}{B}, 2/2 (Morningtide,
// #5592):
//
//	"Changeling (This card is every creature type.)
//	 {B}: This creature gains deathtouch until end of turn. (Any amount of
//	 damage it deals to a creature is enough to destroy that creature.)"
//
// Changeling is every creature type in every zone (CR 702.73a). The
// activation is Battlefly Swarm's: deathtouch on itself until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61b74f82-39f3-4618-b8e7-292058421d1b",
		Name:            "Moonglove Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"deathtouch"}}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Moonglove Changeling — deathtouch until end of turn"),
		}},
	})
}
