package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The three forms of a Corrupted Shapeshifter, as the anchor words its
// "as this creature enters" prompt offers (ADR 0071) — the printed text.
const (
	corruptedShapeshifterFlier  = "3/3 creature with flying"
	corruptedShapeshifterWall   = "2/5 creature with vigilance"
	corruptedShapeshifterDefend = "0/12 creature with defender"
)

// Corrupted Shapeshifter — Creature — Eldrazi Shapeshifter {3}{U}, */*
// (Modern Horizons 3):
//
//	"Devoid (This card has no color.)
//	 As this creature enters, it becomes your choice of a 3/3 creature
//	 with flying, a 2/5 creature with vigilance, or a 0/12 creature with
//	 defender."
//
// The choice is made as the creature enters, so there is never a moment
// on the battlefield with the printed */* (CR 614.12): it is Frostcliff
// Siege's anchor-word choice (ChooseOptionAsEnters) with the form's two
// lines gated on the word, a layer 7b base power and toughness and a layer
// 6 keyword on the creature itself. The answer is the permanent's own, so a
// copy of it chooses again, and +1/+1 counters and anthems still apply to
// the base.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c4cbf1c-95ee-4555-8351-ea962e17ae1f",
		Name:            "Corrupted Shapeshifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		AsEnters: effects.ChooseOptionAsEnters("Corrupted Shapeshifter",
			corruptedShapeshifterFlier, corruptedShapeshifterWall, corruptedShapeshifterDefend),
		Static: []game.StaticAbility{
			effects.StaticWhenChosen(corruptedShapeshifterFlier, effects.GapB53SelfBasePT(3, 3, "This creature is a 3/3 creature with flying.")),
			effects.StaticWhenChosen(corruptedShapeshifterFlier, effects.GapB53SelfKeyword("flying")),
			effects.StaticWhenChosen(corruptedShapeshifterWall, effects.GapB53SelfBasePT(2, 5, "This creature is a 2/5 creature with vigilance.")),
			effects.StaticWhenChosen(corruptedShapeshifterWall, effects.GapB53SelfKeyword("vigilance")),
			effects.StaticWhenChosen(corruptedShapeshifterDefend, effects.GapB53SelfBasePT(0, 12, "This creature is a 0/12 creature with defender.")),
			effects.StaticWhenChosen(corruptedShapeshifterDefend, effects.GapB53SelfKeyword("defender")),
		},
	})
}
