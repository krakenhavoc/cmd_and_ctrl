package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Betrayal of Flesh — Instant {5}{B}:
//
//	"Choose one —
//	 • Destroy target creature.
//	 • Return target creature card from your graveyard to the
//	   battlefield.
//	 Entwine—Sacrifice three lands. (Choose both if you pay the entwine
//	 cost.)"
//
// Entwine (CR 702.42, #3866) with a non-mana cost: the three lands are
// named on the cast (sacrifice_ids) and sacrificed with the spell on the
// stack, as every sacrifice additional cost is (CR 601.2h). Entwined,
// the destroy runs first (CR 702.42b), so the creature it destroys is in
// the graveyard by then — but the card the second bullet returns was
// chosen as a target on the cast, so it is not that one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "43cee9b6-9853-4337-869a-45540c9db10f",
		Name:          "Betrayal of Flesh",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.EntwineSacrifice(3, "three lands", effects.Land())},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Destroy target creature.", effects.TargetCreature("target creature"), effects.DestroyTheModesTarget),
			effects.ModeDoing("Return target creature card from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
		)),
	})
}
