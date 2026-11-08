package effects

// Vengeful Possession — Sorcery {2}{R} (Duskmourn: House of Horror,
// #4466):
//
//	"Gain control of target creature until end of turn. Untap it. It
//	 gains haste until end of turn. You may discard a card. If you do,
//	 draw a card."
//
// Hideous Taskmaster's threaten (threatenAndGrant: control, the untap and
// haste, each for this turn), then Founding of Omashu's "you may discard
// a card. If you do, draw a card" as its own step, so the discard prompt
// comes after the creature has changed hands (CR 608.2c). A card is drawn
// only for a card really discarded. A target that is gone fizzles the
// whole spell, the loot included (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Vengeful Possession"
	const question = "Vengeful Possession — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "1afb9ce6-658e-40e4-a4a2-9a68182a1e73",
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return threatenAndGrant(ctx, id, name, "haste")
			}),
			Instruction(b39MayDiscardThenDraw(1, false, question, drawOneIfDiscarded)),
		),
	})
}
