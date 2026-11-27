package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Indentured Djinn — Creature — Djinn {1}{U}{U}, 4/4 (Mercadian Masques, S82
// OLD, #5619):
//
//	"Flying
//	 When this creature enters, each other player may draw up to three cards."
//
// Shah of Naar Isle's draw-up-to-three for the other players: each chooses none
// to three as the trigger resolves, in turn order after the Djinn's controller,
// and a player a draw limit stops drawing is not asked (#4016).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ccb3bca-73da-450d-a9d0-61d3e2bd05a5",
		Name:            "Indentured Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Indentured Djinn — each other player may draw up to three cards",
				indenturedDjinnEachOtherMayDraw),
		},
	})
}

// indenturedDjinnDrawOptions are the answers to "you may draw up to three cards".
var indenturedDjinnDrawOptions = []game.ChoiceOption{
	{Label: "Draw no cards"},
	{Label: "Draw one card"},
	{Label: "Draw two cards"},
	{Label: "Draw three cards"},
}

// indenturedDjinnEachOtherMayDraw is Indentured Djinn's "each other
// player may draw up to three cards", the choices made in turn order after the
// Djinn's controller (CR 101.4) and every draw made once all have chosen. It is
// Shah of Naar Isle's shape under its own prompt and key.
func indenturedDjinnEachOtherMayDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return indenturedDjinnAskOthers(ctx, effects.OpponentsInTurnOrderAfter(ctx, item.Controller), nil)
}

// indenturedDjinnAskOthers asks each of `others` in order how many cards to draw, then
// draws them all. `chosen` holds the answers so far and is copied, never shared.
func indenturedDjinnAskOthers(ctx *effects.Context, others []uuid.UUID, chosen []int) error {
	if len(chosen) == len(others) {
		for i, p := range others {
			if err := (effects.DrawCards{Player: p, N: chosen[i]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	// CR 121.2b: a draw limit caps how many the player may choose to draw, and
	// one who can't draw at all isn't asked (#4016).
	who := others[len(chosen)]
	ceiling := ctx.Game.DrawChoiceCeilingForEffect(who, len(indenturedDjinnDrawOptions)-1)
	if ceiling <= 0 {
		return indenturedDjinnAskOthers(ctx, others, append(append([]int(nil), chosen...), 0))
	}
	return effects.PickOption{
		Player:   who,
		Question: "Indentured Djinn — you may draw up to three cards",
		Options:  indenturedDjinnDrawOptions[:ceiling+1],
		ThenKey:  indenturedDjinnDrawAnswer,
		Carry:    effects.PackIDLists(others, effects.IntsToCarry(chosen)),
	}.Apply(ctx)
}

var indenturedDjinnDrawAnswer game.OptionPickThen

func init() {
	indenturedDjinnDrawAnswer = effects.OptionPickThen("option-pick/indentured-djinn-draw", func(ctx *effects.Context, r game.OptionPicked) error {
		others, answers := effects.UnpackIDPair(r.Carry)
		index := r.Index
		if index < 0 {
			index = 0
		}
		return indenturedDjinnAskOthers(ctx, others, append(effects.CarryToInts(answers), index))
	})
}
