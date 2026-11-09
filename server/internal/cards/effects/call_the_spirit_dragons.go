package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call the Spirit Dragons — Enchantment {W}{U}{B}{R}{G} (Tarkir:
// Dragonstorm, #4463):
//
//	"Dragons you control have indestructible.
//	 At the beginning of your upkeep, for each color, put a +1/+1 counter
//	 on a Dragon you control of that color. If you put +1/+1 counters on
//	 five Dragons this way, you win the game."
//
// The static is a tribal keyword grant over your Dragons. The upkeep
// trigger walks the five colours in WUBRG order and, for each, asks
// which Dragon of that colour gets the counter (a chained choose_cards
// prompt per colour with a choice to make, std_b11_helpers.go's
// sb11CallStep): a colour with no Dragon of it is skipped, and one with a
// single candidate needs no question. A multicolored Dragon may be the
// pick for several colours and gets a counter for each; the win needs
// counters on five DIFFERENT Dragons (the ruling), so the picks carried
// down the chain are the distinct Dragons counted so far.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ceb23f5-abb1-4569-a1e4-1eed9a9babcf",
		Name:         "Call the Spirit Dragons",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Dragon"}, YoursOnly: true}, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Call the Spirit Dragons — for each color, put a +1/+1 counter on a Dragon you control of that color",
				func(g *game.Game, item *game.StackItem) error {
					return sb11CallStep(g, item.Controller, item.SourceCardID, 0, nil)
				}),
		},
	})
}
