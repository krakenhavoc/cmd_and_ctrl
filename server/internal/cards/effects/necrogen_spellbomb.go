package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrogen Spellbomb — Artifact {1} (Mirrodin, #5606):
//
//	"{B}, Sacrifice this artifact: Target player discards a card.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The first row is Liliana Vess's +1 on a Spellbomb: the target player
// chooses and discards one card as it resolves (CR 701.9a), and a
// player with an empty hand discards nothing.
// The second row is Aether Spellbomb's: {1}, sacrifice it, draw a card
// (mrdB5SpellbombDraw). The sacrifice is part of the cost, so the
// Spellbomb is gone before either row resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34993bd9-53bf-46b2-bc20-b6daf82382d9",
		Name:         "Necrogen Spellbomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}, Sacrifice this artifact: Target player discards a card.",
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
				Cost:    Plus(ManaCost("{B}"), SacrificeThis()),
				Targets: TargetPlayer("target player"),
				Effect: onTheTargetPlayer(func(ctx *Context, p uuid.UUID) error {
					return discardOne(ctx.Game, ctx.Item, p)
				}),
			},
			mrdB5SpellbombDraw(),
		},
	})
}
