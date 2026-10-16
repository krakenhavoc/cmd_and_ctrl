package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windreaver — Creature — Elemental {3}{W}{U}, 1/3:
//
//	"Flying
//	 {W}: This creature gains vigilance until end of turn.
//	 {W}: This creature gets +0/+1 until end of turn.
//	 {U}: Switch this creature's power and toughness until end of turn.
//	 {U}: Return this creature to its owner's hand."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6b6c27e-8028-49b1-90c6-ae0907159f85",
		Name:            "Windreaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: This creature gains vigilance until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"vigilance"}, Label: "Windreaver — vigilance"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{W}: This creature gets +0/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:    ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Toughness: 1, Label: "Windreaver — +0/+1"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{U}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{U}"),
				Effect:  switchThisPT("Windreaver — switch its power and toughness"),
			},
			{
				Label:   "{U}: Return this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{U}"),
				Effect:  returnThisPermanentToOwnersHand,
			},
		},
	})
}
