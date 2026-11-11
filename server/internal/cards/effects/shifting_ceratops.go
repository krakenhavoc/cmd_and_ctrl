package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shifting Ceratops — Creature — Dinosaur {2}{G}{G}, 5/4 (Lost Caverns of
// Ixalan Commander):
//
//	"This spell can't be countered.
//	 Protection from blue
//	 {G}: This creature gains your choice of reach, trample, or haste until
//	 end of turn."
//
// The choice is made as the ability resolves (CR 608.2d), so the same
// activation cannot be answered by picking after seeing a response; the
// Ceratops rides the prompt as data, and a Ceratops that left does
// nothing.
func init() {
	Register(Spec{
		OracleID:        "6cac6f76-f3c6-4b00-a468-166b07853401",
		Name:            "Shifting Ceratops",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{game.ProtectionFromColor("U")},
		Activated: []ActivatedAbility{{
			Label:   "{G}: This creature gains your choice of reach, trample, or haste until end of turn.",
			Cost:    ManaCost("{G}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  shiftingCeratopsChoose,
		}},
	})
}

var shiftingCeratopsKeywords = []string{"reach", "trample", "haste"}

func shiftingCeratopsChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	options := make([]game.ChoiceOption, len(shiftingCeratopsKeywords))
	for i, kw := range shiftingCeratopsKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Shifting Ceratops — gains your choice of reach, trample, or haste until end of turn",
		Options:  options,
		ThenKey:  shiftingCeratopsKeyword,
		Carry:    []uuid.UUID{item.SourceCardID},
	}.Apply(ctx)
}

var shiftingCeratopsKeyword = OptionPickThen("option-pick/shifting-ceratops-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(shiftingCeratopsKeywords) || len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
		return nil
	}
	if !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{shiftingCeratopsKeywords[r.Index]},
		Label:    "Shifting Ceratops — " + shiftingCeratopsKeywords[r.Index],
	}.Apply(ctx)
})
