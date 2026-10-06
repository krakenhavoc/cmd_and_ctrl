package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agonizing Remorse — Sorcery {1}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it or a card from their graveyard. Exile that card. You lose 1
//	 life."
//
// The revealed-hand pick with an exile and the graveyard as well
// (#2115, ADR 0116's 2026-10-05 amendment). The whole table sees the
// hand (CR 701.20a). The candidates are the hand's nonland cards and
// every card in that player's graveyard, lands included (its
// 2020-01-24 ruling). One must be chosen if any exists, from either
// place (the same ruling), and it is exiled, which is not a discard
// (CR 701.9a).
//
// With no nonland card in the hand and an empty graveyard nothing is
// exiled, and you still lose 1 life (CR 609.3, its ruling). The life
// loss is printed after the exile and runs on the next line, before the
// pick is answered; state-based actions wait for the answer (#1289), so
// a caster at 1 life loses only after the exile, as printed. No player
// can act between the choice and the exile (its ruling): the prompt
// blocks the table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1267dfda-eb1a-4963-9fe3-fa619d924d7a",
		Name:         "Agonizing Remorse",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player:        TargetedPlayer(ctx),
				Filter:        Nonland(),
				Label:         "nonland card",
				Exile:         true,
				FromGraveyard: true,
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
		},
	})
}
