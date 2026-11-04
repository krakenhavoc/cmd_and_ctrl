package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b23_helpers.go — shared pieces for the twenty-third Secrets of
// Strixhaven batch (#4456).

// socB23DifferentlyNamedLands is "the number of differently named lands you
// control": each land counted once, however many share its English name.
func socB23DifferentlyNamedLands(ctx *Context) int {
	names := map[string]bool{}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsLand() {
			names[c.Name] = true
		}
	}
	return len(names)
}

// socB23MayTapOrUntap is "you may tap or untap target creature" as the
// spell resolves: the caster picks tap it, untap it or leave it, and the
// answer is applied to the target carried on the prompt.
func socB23MayTapOrUntap(ctx *Context, cardName string, target uuid.UUID) error {
	return PickOption{
		Question: cardName + " — you may tap or untap " + cardNameFor(ctx.Game, target),
		Options: []game.ChoiceOption{
			{Label: "Tap it", Cards: []uuid.UUID{target}},
			{Label: "Untap it", Cards: []uuid.UUID{target}},
			{Label: "Leave it as it is", Cards: []uuid.UUID{target}},
		},
		ThenKey: socB23TapOrUntapAnswer,
		Carry:   []uuid.UUID{target},
	}.Apply(ctx)
}

// socB23TapOrUntapAnswer is the answer to socB23MayTapOrUntap, keyed: the
// creature rides the prompt as its one carried ID. An on-disk key: never
// renamed, never reused.
var socB23TapOrUntapAnswer = OptionPickThen("option-pick/rejoinder-tap-or-untap", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	switch r.Index {
	case 0:
		return TapTarget{Target: r.Carry[0]}.Apply(ctx)
	case 1:
		return UntapTarget{Target: r.Carry[0]}.Apply(ctx)
	default:
		return nil
	}
})
