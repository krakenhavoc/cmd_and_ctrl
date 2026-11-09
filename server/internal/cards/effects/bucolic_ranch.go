package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bucolic Ranch — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Mount
//	 spell.
//	 {3}, {T}: Look at the top card of your library. If it's a Mount card,
//	 you may reveal it and put it into your hand. If you don't put it into
//	 your hand, you may put it on the bottom of your library."
//
// The look is private; the answer is one pick over what the rules allow:
// reveal it into your hand (only when it is a Mount card, read as it sits
// in the library), put it on the bottom, or leave it on top (CR 608.2d).
//
// No simplification.
func init() {
	const name = "Bucolic Ranch"
	Register(Spec{
		OracleID:     "8f5902bf-4bc4-4d0c-84ea-a425307a4eb2",
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{ManaRestrictCast, ManaRestrictSubtype("Mount")},
				Label:        "{T}: Add one mana of any color. Spend this mana only to cast a Mount spell.",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Look at the top card of your library. If it's a Mount card, you may reveal it and put it into your hand. If you don't put it into your hand, you may put it on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect:  bucolicRanchLook,
		}},
	})
}

const (
	bucolicRanchTake   = "Reveal it and put it into your hand"
	bucolicRanchBottom = "Put it on the bottom of your library"
	bucolicRanchLeave  = "Leave it on top"
)

// bucolicRanchLook looks at the top card and asks what to do with it.
func bucolicRanchLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	top := g.LookAtTopOfLibraryForEffect(you, 1)
	if len(top) == 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: bucolicRanchBottom}, {Label: bucolicRanchLeave}}
	if c, ok := g.LookupCardForEffect(top[0]); ok && !c.IsToken() && c.HasSubtype(g.TextReaderForEffect(item).TextSubtype("Mount")) {
		options = append([]game.ChoiceOption{{Label: bucolicRanchTake}}, options...)
	}
	return PickOption{
		Question: "Bucolic Ranch — what do you do with the top card of your library?",
		Options:  options,
		ThenKey:  bucolicRanchAnswer,
		Carry:    []uuid.UUID{top[0]},
	}.Apply(ctx)
}

// bucolicRanchAnswer carries out the pick. An on-disk key: never renamed,
// never reused.
var bucolicRanchAnswer = OptionPickThen("option-pick/bucolic-ranch-top-card", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	g, id := ctx.Game, r.Carry[0]
	switch r.Option.Label {
	case bucolicRanchTake:
		g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Reason: "Bucolic Ranch — reveal the Mount card", Cards: []uuid.UUID{id}})
		g.TakeFromLibraryToHandForEffect(r.Chooser, []uuid.UUID{id})
	case bucolicRanchBottom:
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{id})
	}
	return nil
})
