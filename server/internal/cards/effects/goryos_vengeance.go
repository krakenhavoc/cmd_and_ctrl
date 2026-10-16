package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goryo's Vengeance — Instant — Arcane {1}{B}:
//
//	"Return target legendary creature card from your graveyard to the
//	battlefield. That creature gains haste. Exile it at the beginning of
//	the next end step.
//	Splice onto Arcane {2}{B}"
//
// Whip of Erebos's return without its leave-the-battlefield redirect:
// the creature returns, gains haste, and the next end step's delayed
// trigger exiles it if it is still the object that returned.
func init() {
	Register(Spec{
		OracleID:     "3ea3c681-f0ad-45a2-bc25-2813efd8d560",
		Name:         "Goryo's Vengeance",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{2}{B}"),
		Targets:      TargetCardInGraveyard("target legendary creature card from your graveyard", Creature(), Legendary(), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			g := ctx.Game
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
				return err
			}
			if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"haste"},
				Label:    "Goryo's Vengeance — the returned creature has haste",
			}).Apply(ctx); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				Label: "Goryo's Vengeance — exile the returned creature",
				Cards: []uuid.UUID{t.ID},
				Body:  exileListedCardsBody,
			}.Apply(ctx)
		},
	})
}
