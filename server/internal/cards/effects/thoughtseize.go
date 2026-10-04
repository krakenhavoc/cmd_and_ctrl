package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtseize — Sorcery {B}:
//
//	"Target player reveals their hand. You choose a nonland card from
//	 it. That player discards that card. You lose 2 life."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), only nonland cards may be chosen — an MDFC with a
// spell front is one (CR 712.8a) — and the server refuses any other
// pick. A hand with no nonland card is revealed and nothing is
// discarded (CR 609.3).
//
// The life loss is printed after the discard and runs on the next line
// here, before the pick is answered. State-based actions and triggers
// wait for the answer (#1289), so a caster at 2 life still loses only
// after the discard. The loss happens with no nonland card to take
// too: the 2020-08-07 ruling.
//
// Thoughtseize can target its caster; the caster then reveals their
// own hand to the table and picks from it.
func init() {
	Register(Spec{
		OracleID:     "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		Name:         "Thoughtseize",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
