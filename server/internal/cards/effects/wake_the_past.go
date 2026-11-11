package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wake the Past — Sorcery {5}{R}{W} (Foundations Commander):
//
//	"Return all artifact cards from your graveyard to the battlefield.
//	 They gain haste until end of turn."
//
// Splendid Reclamation's together-return over the artifact cards of the
// caster's graveyard; the haste is the return's tail, told which cards
// entered (Pull's haste tail), so a card that failed to come back gains
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34143cfa-aa48-481c-af42-1240af174b80",
		Name:         "Wake the Past",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets:  graveyardCardIDs(ctx, ctx.Controller(), game.Card.IsArtifact),
				ThenTail: cmdrB9HasteTail,
			}.Apply(ctx)
		},
	})
}

// cmdrB9HasteTail is "they gain haste until end of turn" over the cards
// a return settled with. An on-disk key.
var cmdrB9HasteTail = TogetherTail("cmdr-b9/gain-haste-until-eot", func(ctx *Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Wake the Past"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
