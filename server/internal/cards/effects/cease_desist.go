package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cease // Desist — split card, oracle ba4d644f-1931-4fc4-aed5-681a476a5a58:
//
//	Cease — Instant {1}{B/G}: "Exile up to two target cards from a
//	        single graveyard. Target player gains 2 life and draws a
//	        card."
//	Desist — Sorcery {4}{G/W}{G/W}: "Destroy all artifacts and
//	         enchantments."
//
// #1807, ADR 0106 §5. Each half is an ADR 0034 face (ADR 0103): Cease
// under the bare oracle ID, Desist under "#1". Only one half can be
// cast (CR 709.3; the card has no fuse).
//
// Cease has two target clauses: the cards, which must share a
// graveyard, then the player. Each is read from its own slot. The
// player is not tied to the graveyard — any player may gain the life,
// whoever's graveyard was hit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba4d644f-1931-4fc4-aed5-681a476a5a58",
		Name:         "Cease",
		Completeness: CompletenessFull,
		Targets: Clauses(
			upToNCardsFromASingleGraveyard(2),
			TargetPlayer("target player"),
		),
		// The second sentence runs from the exile's continuation, so it
		// follows the exile even when a commander card among the targets
		// pauses it on its owner's CR 903.9 choice (CR 608.2c: in order).
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetCardsThen(ctx, func(ctx *Context, _ []uuid.UUID, _ map[uuid.UUID]bool) error {
				p, ok := ctx.ClauseTarget(1)
				if !ok || p.Kind != game.TargetPlayer {
					return nil
				}
				if err := (GainLife{Player: p.ID, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: p.ID, N: 1}.Apply(ctx)
			})
		},
	})
	Register(Spec{
		OracleID:     "ba4d644f-1931-4fc4-aed5-681a476a5a58#1",
		Name:         "Desist",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Or(Artifact(), Enchantment())}.Apply(ctx)
		},
	})
}
