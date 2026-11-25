package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b315_helpers.go — shared pieces of OLD batch 315, Battle for
// Zendikar (S82, #5547).

// S82OldB315WarleaderAsk is Veteran Warleader's "This creature gains your
// choice of first strike, vigilance, or trample until end of turn": the
// choice is made as the ability resolves, and the keyword goes to the
// creature the ability came from, if it is still there.
func S82OldB315WarleaderAsk(g *game.Game, item *game.StackItem) error {
	return PickOption{
		Question: "Veteran Warleader — gains which until end of turn?",
		Options:  []game.ChoiceOption{{Label: "First strike"}, {Label: "Vigilance"}, {Label: "Trample"}},
		ThenKey:  s82OldB315WarleaderKeyword,
		Carry:    []uuid.UUID{item.SourceCardID},
	}.Apply(NewContext(g, item))
}

// s82OldB315WarleaderKeyword is the keyword chosen, keyed (#4253): the
// creature rides the prompt as data. An on-disk key.
var s82OldB315WarleaderKeyword = OptionPickThen("option-pick/s82-old-b315-veteran-warleader-keyword", func(ctx *Context, r game.OptionPicked) error {
	keywords := []string{"first strike", "vigilance", "trample"}
	if r.Index < 0 || r.Index >= len(keywords) || len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{keywords[r.Index]},
		Label:    "Veteran Warleader — " + keywords[r.Index],
	}.Apply(ctx)
})

// S82OldB315ColorlessCreatureMV7 is the search clause of Conduit of Ruin: "a
// colorless creature card with mana value 7 or greater".
func S82OldB315ColorlessCreatureMV7(c game.Card) bool {
	return c.IsCreature() && c.IsColorless() && c.ManaValue() >= 7
}
