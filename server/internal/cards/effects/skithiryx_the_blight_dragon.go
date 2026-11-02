package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skithiryx, the Blight Dragon — Legendary Creature — Phyrexian Dragon
// Skeleton {3}{B}{B}, 4/4:
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)
//	 {B}: Skithiryx gains haste until end of turn.
//	 {B}{B}: Regenerate Skithiryx."
//
// Two printed keywords and two activated abilities, each a standing
// shape: Endling's self grant and Albino Troll's regeneration shield
// (CR 701.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "daf6c421-e7f7-4fc6-967c-65f4ab96fcfd",
		Name:            "Skithiryx, the Blight Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: Skithiryx gains haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{B}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("haste", "Skithiryx, the Blight Dragon — haste until end of turn"),
			},
			{
				Label:   "{B}{B}: Regenerate Skithiryx.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{B}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
