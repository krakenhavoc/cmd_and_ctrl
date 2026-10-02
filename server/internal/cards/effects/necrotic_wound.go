package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrotic Wound — Instant {B}:
//
//	"Undergrowth — Target creature gets -X/-X until end of turn, where X
//	 is the number of creature cards in your graveyard. If that creature
//	 would die this turn, exile it instead."
//
// Undergrowth is an ability word: nothing but the count. X is counted
// as the spell resolves (CR 608.2h) and locked into the shrink then
// (CR 611.2c), so creature cards that reach the graveyard later do not
// grow it. The replacement is the spell's (ADR 0108 §1): the target is
// marked even when X is 0, and a creature the shrink puts into the
// graveyard (CR 704.5f) dies and is exiled instead.
//
// No simplifications.
func init() {
	creatureCards := g2CardsInYourGraveyard(game.Card.IsCreature)
	Register(Spec{
		OracleID:     "a30159ae-f6a6-4e29-bca2-769d3657d310",
		Name:         "Necrotic Wound",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := creatureCards(item, ctx)
			if err := (BoostUntilEOT{Target: id, Power: -x, Toughness: -x}).Apply(ctx); err != nil {
				return err
			}
			return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
		},
	})
}
