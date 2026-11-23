package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tune Up — Sorcery {3}{W} (Aetherdrift, #4464):
//
//	"Return target artifact card from your graveyard to the battlefield.
//	 If it's a Vehicle, it becomes an artifact creature."
//
// It Doesn't Add Up's return with the second sentence as the entry's tail
// (ReturnFromGraveyardToBattlefieldKeyedForEffect), so it runs on the
// permanent that came back (CR 400.7j) both when the artifact simply
// enters and after an entry that first asks something. "If it's a
// Vehicle" is read off that permanent on the battlefield. The animation
// has no duration: BecomeArtifactCreature is the indefinite form, so the
// Vehicle stays an artifact creature until it leaves (CR 400.7), with
// its printed power and toughness. A card that left the graveyard in
// response is no longer a legal target and the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaa14260-2537-4668-8517-c825118777cc",
		Name:         "Tune Up",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(target, ctx.Controller(), false, ctx.Tail(tuneUpVehicleBecomesACreature))
		},
	})
}

// tuneUpVehicleBecomesACreature is the second sentence, told the permanent
// that entered (uuid.Nil when nothing did). An on-disk key.
var tuneUpVehicleBecomesACreature = effects.EnteredTail("tune-up/vehicle-becomes-artifact-creature",
	func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
		if entered == uuid.Nil || !effects.B15OnBattlefield(ctx.Game, entered) {
			return nil
		}
		if c, ok := ctx.Game.LookupCardForEffect(entered); !ok || !c.HasSubtype("Vehicle") {
			return nil
		}
		return effects.BecomeArtifactCreature{Target: entered, Label: "Tune Up — becomes an artifact creature"}.Apply(ctx)
	})
