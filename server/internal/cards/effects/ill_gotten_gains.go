package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ill-Gotten Gains — Sorcery {2}{B}{B}:
//
//	"Exile Ill-Gotten Gains. Each player discards their hand, then
//	 returns up to three cards from their graveyard to their hand."
//
// The spell exiles ITSELF first, exactly as Finale of Revelation's last
// sentence does: it is still on the stack while it resolves, and the
// engine sees the card has already left and skips the graveyard
// routing. So it is not in the graveyard for anyone to return.
//
// Every hand is discarded before anyone chooses, so the cards just
// discarded are candidates. Then each player is asked, about THEIR OWN
// graveyard (CR 101.4: asked in turn order, answered in secret), for up
// to three cards. The questions are one run (ChooseCardsRunThenKeyedForEffect)
// so nothing moves until the last player has answered, which is the
// "returns" being simultaneous; a player whose graveyard is empty is
// not asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c88eff74-def5-42ea-873c-b3b479f6fe18",
		Name:         "Ill-Gotten Gains",
		Completeness: CompletenessFull,
		// The graveyard picks wait for the prompts the discards ask (a replacement offer such as Library of Leng's; CR
		// 608.2c, #4155) and read the graveyards when they run.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if err := (ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range tablePlayers(ctx) {
					if _, err := discardWholeHand(ctx.Game, id); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(illGottenGainsReturn),
		),
	})
}

// illGottenGainsReturn is "each player may return up to three cards from
// their graveyard to their hand".
func illGottenGainsReturn(ctx *Context) error {
	players := tablePlayers(ctx)
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		cards := graveyardCardsOfPlayer(ctx.Game, id)
		if len(cards) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "Ill-Gotten Gains — return up to three cards from your graveyard to your hand",
			Cards:      cards,
			Min:        0,
			Max:        3,
			Zone:       game.ZoneGraveyard,
		})
	}
	// Keyed (#4290): the players, in ask order, ride the run as its carry.
	_, err := ctx.Game.ChooseCardsRunThenKeyedForEffect(legs, illGottenGainsReturned, ctx.Source(), ctx.RunData(players))
	return err
}

// illGottenGainsReturned is the cards the players chose returning to their
// hands, once every player has chosen, keyed (#4290).
var illGottenGainsReturned = RunThen("prompt-run/ill-gotten-gains-return", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	for _, id := range r.Carry {
		for _, card := range picks.By(id) {
			if err := (ReturnFromGraveyard{Target: card, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
