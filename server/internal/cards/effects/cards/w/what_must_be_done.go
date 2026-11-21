package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// What Must Be Done — Sorcery {3}{W}{W}:
//
//	"Choose one —
//	 • Let the World Burn — Destroy all artifacts and creatures.
//	 • Release Juno — Return target historic permanent card from your
//	   graveyard to the battlefield. It enters with two additional +1/+1
//	   counters on it if it's a creature. (Artifacts, legendaries, and
//	   Sagas are historic.)"
//
// Let the World Burn is Ultima's sweep: one simultaneous destroy over
// artifacts and creatures, so an indestructible permanent survives. Release
// Juno returns the first legal target under its controller's control with
// the counters as the entry's own (CR 614.1c, Heroic Return's shape),
// given when the card is a creature as it would exist on the battlefield
// (CR 614.12). Historic is CR 700.6.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd0df0a8-c2ec-4d1f-9ada-cc91f7a20e1a",
		Name:         "What Must Be Done",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Let the World Burn — Destroy all artifacts and creatures.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DestroyAllMatching{Match: effects.Or(effects.Artifact(), effects.Creature())}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
			effects.ModeDoing("Release Juno — Return target historic permanent card from your graveyard to the battlefield. It enters with two additional +1/+1 counters on it if it's a creature.",
				effects.TargetCardInGraveyard("target historic permanent card from your graveyard", effects.YouOwn(), effects.Permanent(), effects.B14Historic()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					var id uuid.UUID
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							id = t.ID
							break
						}
					}
					if id == uuid.Nil {
						return nil
					}
					if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					opts := game.ZoneEntryOptions{
						Controller: ctx.Controller(),
						EntersWithCountersIfType: []game.EntryCountersIfType{
							{CardType: "creature", Counter: game.CounterPlusOne, N: 2},
						},
					}
					return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{id}, opts, game.TailRef{}, 0)
				}),
		),
	})
}
