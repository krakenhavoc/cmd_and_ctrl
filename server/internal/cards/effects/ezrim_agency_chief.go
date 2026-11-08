package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ezrim, Agency Chief — Legendary Creature — Archon Detective
// {1}{W}{W}{U}{U}, 5/5 (Murders at Karlov Manor, #4469):
//
//	"Flying
//	 When Ezrim enters, investigate twice. (To investigate, create a Clue
//	 token. It's an artifact with "{2}, Sacrifice this token: Draw a
//	 card.")
//	 {1}, Sacrifice an artifact: Ezrim gains your choice of vigilance,
//	 lifelink, or hexproof until end of turn."
//
// Investigating twice is two Clues (Detective's Satchel). The activated
// ability sacrifices any artifact as its cost (Arcbound Ravager's
// sacrificeSpec), a Clue included, and asks for the keyword as it
// resolves (Gideon Blackblade's +1, CR 608.2c). The keyword lands on
// Ezrim himself; if he has left by then, the grant does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ddcd8a0a-7ac3-4400-a16f-8bb5aa50994e",
		Name:            "Ezrim, Agency Chief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ezrim, Agency Chief — investigate twice", Do(CreateToken{Template: ClueToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: Ezrim gains your choice of vigilance, lifelink, or hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{1}"), game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())}),
			Effect:  ezrimChooseAKeyword,
		}},
	})
}

// ezrimKeywords are the three keywords the ability offers, in printed
// order.
var ezrimKeywords = []string{"vigilance", "lifelink", "hexproof"}

// ezrimChooseAKeyword asks the controller which keyword Ezrim gains.
func ezrimChooseAKeyword(g *game.Game, item *game.StackItem) error {
	options := make([]game.ChoiceOption, len(ezrimKeywords))
	for i, kw := range ezrimKeywords {
		options[i] = game.ChoiceOption{Label: "Ezrim gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Ezrim, Agency Chief — gains your choice of vigilance, lifelink, or hexproof until end of turn",
		Options:  options,
		ThenKey:  ezrimKeywordChosen,
		WithItem: true,
	}.Apply(NewContext(g, item))
}

// ezrimKeywordChosen grants the chosen keyword to Ezrim, the ability's
// source, until end of turn. The Context is rebuilt from the resolving
// item (WithItem), so its source stamp still refuses an Ezrim that left
// and came back as a new object (CR 400.7).
var ezrimKeywordChosen = OptionPickThen("option-pick/ezrim-agency-chief-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(ezrimKeywords) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   ctx.Source(),
		Keywords: []string{ezrimKeywords[r.Index]},
		Label:    "Ezrim, Agency Chief — gains " + ezrimKeywords[r.Index] + " until end of turn",
	}.Apply(ctx)
})
