package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverquill Pledgemage — Creature — Vampire Cleric {1}{W/B}{W/B}, 3/1
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, this
//	 creature gains your choice of flying or lifelink until end of turn."
//
// The choice is made as the trigger resolves (Wingnut's keyword prompt) and
// the keyword is granted only while the Pledgemage is still the permanent
// the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff72b900-ce3c-4cc4-a50e-e501e17631ca",
		Name:         "Silverquill Pledgemage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Silverquill Pledgemage — gains your choice of flying or lifelink until end of turn (magecraft)",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.PickOption{
						Question: "Silverquill Pledgemage gains which until end of turn?",
						Options:  []game.ChoiceOption{{Label: "Flying"}, {Label: "Lifelink"}},
						ThenKey:  silverquillPledgemageKeywordChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// silverquillPledgemageKeywords are the two choices, in the order the question lists them.
var silverquillPledgemageKeywords = []string{"flying", "lifelink"}

// silverquillPledgemageKeywordChosen grants the chosen keyword to the Pledgemage
// (the prompt's source) until end of turn. An on-disk key, never renamed.
var silverquillPledgemageKeywordChosen = effects.OptionPickThen("option-pick/silverquill-pledgemage-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(silverquillPledgemageKeywords) || !effects.OnBattlefield(ctx.Game, r.Source) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   r.Source,
		Keywords: []string{silverquillPledgemageKeywords[r.Index]},
		Label:    "Silverquill Pledgemage — " + silverquillPledgemageKeywords[r.Index] + " until end of turn",
	}.Apply(ctx)
})
