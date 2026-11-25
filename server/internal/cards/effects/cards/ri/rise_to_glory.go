package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise to Glory — Sorcery {3}{W}{B}:
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to the battlefield.
//	 • Return target Aura card from your graveyard to the battlefield."
//
// The modes run in the order printed. A returned Aura is not cast, so it
// chooses what it enchants as it enters (CR 303.4f), and may choose the
// creature the first mode returned.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a41297c4-5c2d-475d-b1f5-439df30fbad7",
		Name:         "Rise to Glory",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target creature card from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			effects.ModeDoing("Return target Aura card from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target Aura card from your graveyard", effects.HasSubtype("Aura"), effects.YouOwn()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
		),
	})
}
