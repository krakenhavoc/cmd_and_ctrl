package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Dread Marn — Instant {2}{B}:
//
//	"Create X 2/2 black Zombie Berserker creature tokens, where X is the
//	 number of nontoken creatures that died this turn.
//	 Foretell {B}"
//
// X is counted as the spell resolves, over the turn's event log: every
// creature that went from the battlefield to a graveyard this turn
// (judged on its last-known types), anyone's, less the tokens. A token
// is told apart by the card the event names: a token that died has
// ceased to exist (CR 704.5d), and one still findable says so itself.
// A card that died twice this turn was two creatures and counts twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0998768a-1632-4da1-8310-352582501c4b",
		Name:         "Rise of the Dread Marn",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Foretell("{B}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/2 black Zombie Berserker"), N: nontokenCreaturesDiedThisTurn(ctx.Game)}.Apply(ctx)
		},
	})
}

// nontokenCreaturesDiedThisTurn counts the nontoken creatures that died
// this turn, any controller.
func nontokenCreaturesDiedThisTurn(g *game.Game) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.IsToken() || !leftAsType(ev, c, "creature") {
			continue
		}
		n++
	}
	return n
}
