package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cleansing — Sorcery {W}{W}{W}:
//
//	"For each land, destroy that land unless any player pays 1 life."
//
// ADR 0187 §2 (#3574): one offer for the whole table, not a prompt per
// land. Every player in APNAP order chooses which lands they will pay 1
// life for, knowing what the earlier players chose (CR 101.4b), then the
// payments are all made, one life payment per land chosen. Two players
// who chose the same land both pay and the land is saved once. The lands
// nobody paid for are destroyed together in one batch (CR 701.8a), so
// regeneration and indestructible apply as they do for Armageddon, and
// every "dies" trigger sees the whole batch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46b83657-1699-4162-a4bb-b792da396fe6",
		Name:         "Cleansing",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return AnyPlayerPaysForEach{
				Objects: func(ctx *Context) []uuid.UUID {
					var lands []uuid.UUID
					for _, c := range MatchingBattlefield(ctx, Land()) {
						lands = append(lands, c.InstanceID)
					}
					return lands
				},
				EachLife: 1,
				Noun:     "land",
				Question: "Cleansing — pay 1 life for each land you want to save?",
				OnUnsaved: func(ctx *Context, unsaved []uuid.UUID) error {
					return ctx.Game.DestroyPermanentsThenForEffect(unsaved, func(*game.Game, []uuid.UUID) error { return nil })
				},
			}.Apply(ctx)
		},
	})
}
