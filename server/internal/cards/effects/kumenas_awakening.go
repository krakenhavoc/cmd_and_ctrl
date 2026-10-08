package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kumena's Awakening — Enchantment {2}{U}{U} (EDHREC rank 12437):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 At the beginning of your upkeep, each player draws a card. If you
//	 have the city's blessing, instead only you draw a card."
//
// A symmetrical Howling Mine that turns into a private card engine
// once the controller's board is wide. The condition is a clause of the
// effect, not an intervening "if" on the trigger, so it is read as the
// ability RESOLVES (CR 608.2): a controller who earned the blessing in
// response to the trigger draws alone.
//
// Each player draws in turn order starting with the controller, the
// order a table draws in when told "each player" (CR 101.4, APNAP).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7a8066e-f8a4-4202-8825-0327d11e58c3",
		Name:            "Kumena's Awakening",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Kumena's Awakening — each player draws a card (only you, with the city's blessing)",
				kumenasAwakeningDraw),
		},
	})
}

func kumenasAwakeningDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if YouHaveTheCitysBlessing(g, item.Controller) {
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	n := len(g.Seats)
	if n == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < n; i++ {
		p := g.Seats[(start+i)%n]
		if p == nil || p.Eliminated {
			continue
		}
		if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
