package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kokusho, the Evening Star — Legendary Creature — Dragon Spirit
// {4}{B}{B}, 5/5 (EDHREC rank 2298):
//
//	"Flying
//	 When Kokusho dies, each opponent loses 5 life. You gain life
//	 equal to the life lost this way."
//
// The Kamigawa Dragon that drains the table on the way out. Flying
// rides PrintedKeywords; the dies trigger (cardDied — the graveyard
// only, so a bounce or an exile is not a death) is Gray Merchant's
// drain at a fixed five: each opponent loses 5 life — life loss, not
// damage — and the controller gains 5 per opponent who lost it,
// fifteen at a full table. The trigger's controller is Kokusho's
// controller as it died, read off the card in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf631c93-b7fb-4e4f-b405-3778c15b1117",
		Name:            "Kokusho, the Evening Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Kokusho, the Evening Star — each opponent loses 5 life, you gain that much", func(g *game.Game, item *game.StackItem) error {
				return b21DrainEachOpponentAndGainTheTotal(g, item, 5)
			}),
		},
	})
}
