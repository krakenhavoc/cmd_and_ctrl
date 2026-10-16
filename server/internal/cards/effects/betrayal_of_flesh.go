package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:      "43cee9b6-9853-4337-869a-45540c9db10f",
		Name:          "Betrayal of Flesh",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{EntwineSacrifice(3, "three lands", Land())},
		Modes: Entwining(ChooseOne(
			ModeDoing("Destroy target creature.", TargetCreature("target creature"), DestroyTheModesTarget),
			ModeDoing("Return target creature card from your graveyard to the battlefield.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
		)),
	})
}
