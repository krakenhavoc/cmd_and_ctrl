package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaretiller — Artifact Creature — Scarecrow {4}, 1/4 (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"Whenever this creature becomes tapped, choose one —
//	 • You may put a land card from your hand onto the battlefield tapped.
//	 • Return target land card from your graveyard to the battlefield tapped."
//
// "Becomes tapped" is Chrome Companion's ThisBecameTapped: the tap event and the
// attack that taps it, so a vigilance attack is not one (CR 603.2e). The two
// bullets are a modal trigger's (CR 700.2): the mode is chosen as the ability
// goes on the stack, and the graveyard bullet is offered only with a land card
// in the graveyard to target. The hand bullet is Arboreal Grazer's put (a put,
// not a play, so the turn's land drop is untouched).
//
// No simplification.
func init() {
	t := effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
		"Scaretiller — choose one: put a land card from your hand onto the battlefield tapped, or return target land card from your graveyard to the battlefield tapped",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("You may put a land card from your hand onto the battlefield tapped.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.MayPutALandFromHandTapped("Scaretiller").Apply(ctx)
			}),
		effects.ModeDoing("Return target land card from your graveyard to the battlefield tapped.",
			effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				tgt, ok := effects.ModeTarget(ctx, occ)
				if !ok || tgt.Kind != game.TargetCard {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: tgt.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e4b32e47-5484-4ee9-bac3-4ba37558a56a",
		Name:         "Scaretiller",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
