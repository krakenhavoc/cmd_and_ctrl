package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uncle's Musings — Sorcery {2}{G}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Converge — Return up to X permanent cards from your graveyard to
//	 your hand, where X is the number of colors of mana spent to cast
//	 this spell.
//	 Exile Uncle's Musings."
//
// Painful Truths' converge count (the distinct colours that paid,
// colourless not one of them) bounds one choose-cards pick over the
// permanent cards in your graveyard: the cards are chosen as the spell
// resolves, not targeted. Zero colours offers nothing. The spell then
// exiles itself instead of going to the graveyard, whether or not any
// card came back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "6de489bf-6922-4f97-8001-efa8cb5a27cb",
		Name:                "Uncle's Musings",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if x := ctx.ColorsSpentCount(); x > 0 {
				if err := (ReturnChosenFromGraveyard{
					Question: "Uncle's Musings — return up to " + strconv.Itoa(x) + " permanent cards from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsPermanent() },
					Min:      0,
					Max:      x,
					ToHand:   true,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			// "Exile Uncle's Musings." — only the spell's own card, not a
			// copy, and not a card that already became a new object.
			if item.IsCopy || sourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
