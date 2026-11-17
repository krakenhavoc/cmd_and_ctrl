package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farideh, Devil's Chosen — Legendary Creature — Tiefling Warlock 3/3
// {2}{U}{R} (Adventures in the Forgotten Realms, #5681):
//
//	"Dark One's Own Luck — Whenever you roll one or more dice, Farideh
//	 gains flying and menace until end of turn. If any of those results
//	 was 10 or higher, draw a card."
//
// One trigger per instruction to roll, however many dice it rolled
// (WheneverYouRollDice). "Those results" are the results of the dice that
// instruction kept, each its natural result plus the instruction's
// modifier (CR 706.2, RolledDice): Diviner's Portent's 4 with six cards
// in hand is a result of 10. An ignored roll never happened (CR 706.6),
// and a planar die triggers the ability but has no numerical result (CR
// 706.7), so it never draws. One card at most, however many dice were 10
// or higher. The draw does not depend on Farideh: she still draws after
// leaving the battlefield, and only the keywords are lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "faf5c18c-2642-4df3-a21f-3b237527e6a6",
		Name:         "Farideh, Devil's Chosen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouRollDice("Farideh, Devil's Chosen — gains flying and menace until end of turn; draw a card if any result was 10 or higher",
				func(g *game.Game, item *game.StackItem, _ int) error {
					ctx := NewContext(g, item)
					if zone := g.FindCardZoneForEffect(item.SourceCardID); zone != nil && zone.Kind == game.ZoneBattlefield && !sourceIsNewObject(g, item) { // #1432
						grant := GrantKeywordUntilEOT{
							Target:   item.SourceCardID,
							Keywords: []string{"flying", "menace"},
							Label:    "Farideh, Devil's Chosen — flying and menace until end of turn",
						}
						if err := grant.Apply(ctx); err != nil {
							return err
						}
					}
					for _, die := range RolledDice(g, item) {
						if die.RollResult() >= 10 {
							return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
						}
					}
					return nil
				}),
		},
	})
}
