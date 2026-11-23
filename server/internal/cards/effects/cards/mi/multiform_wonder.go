package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Multiform Wonder — Artifact Creature — Construct {5}, 3/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}: This creature gains your choice of flying, vigilance, or
//	 lifelink until end of turn.
//	 Pay {E}: This creature gets +2/-2 or -2/+2 until end of turn."
//
// ADR 0129 §2: both rows pay one energy (PayEnergy). The two choices
// are made as each ability RESOLVES, not when it is activated, so they
// are asked then (PickOption for the keyword, MayChoice for the two
// pumps), the way Lunar Avenger and Endling ask theirs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9849c743-9492-4a1b-83bf-1be54614b85f",
		Name:         "Multiform Wonder",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Multiform Wonder", 3),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Pay {E}: This creature gains your choice of flying, vigilance, or lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.PayEnergy(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PickOption{
						Question: "Multiform Wonder gains which until end of turn?",
						Options: []game.ChoiceOption{
							{Label: "Flying"}, {Label: "Vigilance"}, {Label: "Lifelink"},
						},
						ThenKey: multiformWonderKeyword,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "Pay {E}: This creature gets +2/-2 or -2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.PayEnergy(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Multiform Wonder — +2/-2 or -2/+2 until end of turn?",
						YesLabel: "+2/-2",
						NoLabel:  "-2/+2",
						Then:     multiformWonderPump,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// multiformWonderPump is Multiform Wonder's choice: +2/-2 on a yes, -2/+2 on
// a no. Keyed (#4291).
var multiformWonderPump = effects.MayYesNo("confirm/multiform-wonder-pump",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(2, -2, "Multiform Wonder — +2/-2 until end of turn")(ctx.Game, ctx.Item)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(-2, 2, "Multiform Wonder — -2/+2 until end of turn")(ctx.Game, ctx.Item)
	})

// multiformWonderKeyword is the keyword chosen, keyed (#4253): the answer is
// the index into the three options.
var multiformWonderKeyword = effects.OptionPickThen("option-pick/multiform-wonder-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	keywords := []string{"flying", "vigilance", "lifelink"}
	if r.Index < 0 || r.Index >= len(keywords) {
		return nil
	}
	return effects.ThisGainsKeywordUntilEndOfTurn(keywords[r.Index], "Multiform Wonder — "+keywords[r.Index]+" until end of turn")(ctx.Game, ctx.Item)
})
