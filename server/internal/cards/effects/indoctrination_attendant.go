package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indoctrination Attendant — Creature — Phyrexian Cleric {3}{W}, 3/4:
//
//	"Toxic 1
//	 When this creature enters, you may return another permanent you control
//	 to its owner's hand. If you do, create a 1/1 colorless Phyrexian Mite
//	 artifact creature token with toxic 1 and 'This token can't block.'"
//
// The ETB does not target, so the permanent is picked as it resolves (the
// Ambrosia Whiteheart shape). The Mite is the bounce's registered tail, run
// only when the permanent reached a hand (a commander that took the command
// zone instead makes none).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9587bdb3-fb6a-47c0-898d-404bf60fcb6b",
		Name:            "Indoctrination Attendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Indoctrination Attendant — you may return another permanent you control to its owner's hand; if you do, create a Phyrexian Mite", indoctrinationAttendantReturn),
		},
	})
}

func indoctrinationAttendantReturn(g *game.Game, item *game.StackItem) error {
	return s82OneB04ReturnAnotherPermanent(g, item,
		"Indoctrination Attendant — you may return another permanent you control to its owner's hand",
		indoctrinationAttendantPicked)
}

var indoctrinationAttendantPicked = RegisterPermanentsPick("prompt-run/indoctrination-attendant-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			if err := (BounceToHand{Target: id, ThenTail: ctx.ItemTail(indoctrinationAttendantMite)}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// indoctrinationAttendantMite is "if you do, create a Mite". An on-disk key.
var indoctrinationAttendantMite = CardTail("indoctrination-attendant/mite", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return CreateToken{Template: PhyrexianMiteToken(), N: 1}.Apply(ctx)
})
