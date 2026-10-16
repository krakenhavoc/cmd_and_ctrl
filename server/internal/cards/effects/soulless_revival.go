package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulless Revival — Instant — Arcane {1}{B}:
//
//	"Return target creature card from your graveyard to your hand.
//	Splice onto Arcane {1}{B}"
func init() {
	Register(Spec{
		OracleID:     "5f00dcc4-38e7-4dde-96d4-f8d4c08b334a",
		Name:         "Soulless Revival",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{B}"),
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
