package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambassador K'Ehleyr — Legendary Creature — Klingon Human Advisor
// {2}{W}, 2/2 (Star Trek Commander, #4974):
//
//	"Vigilance
//	 Parley — At the beginning of combat on your turn, each player
//	 reveals the top card of their library, then draws a card. For each
//	 nonland card revealed this way, put a +1/+1 counter on Ambassador
//	 K'Ehleyr."
//
// "Parley" is an ability word. Every player still in the game reveals,
// in APNAP order, before anyone draws; then each draws one (so the card
// drawn is the card revealed). A player with an empty library reveals
// nothing, and the draw from it is the CR 704.5b loss as usual. The
// counters go on K'Ehleyr only while she is still the permanent the
// ability came from (CR 400.7).
func init() {
	Register(Spec{
		OracleID:        "ce60a1a4-1e53-4985-a54f-3627dd35c5f2",
		Name:            "Ambassador K'Ehleyr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(ambassadorKEhleyrLabel, ambassadorKEhleyrParley),
		},
	})
}

const ambassadorKEhleyrLabel = "Ambassador K'Ehleyr — each player reveals the top card of their library, then draws a card"

func ambassadorKEhleyrParley(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := apnapPlayers(g)
	nonland := 0
	for _, p := range players {
		for _, id := range g.RevealTopOfLibraryForEffect(p, ctx.Source(), 1, ambassadorKEhleyrLabel) {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland++
			}
		}
	}
	for _, p := range players {
		if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if nonland == 0 || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: nonland}.Apply(ctx)
}
