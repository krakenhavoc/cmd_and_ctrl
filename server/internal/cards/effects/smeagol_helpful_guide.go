package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sméagol, Helpful Guide — Legendary Creature — Halfling Horror
// {1}{B}{G}, 4/2:
//
//	"At the beginning of your end step, if a creature died under your
//	 control this turn, the Ring tempts you.
//	 Whenever the Ring tempts you, target opponent reveals cards from
//	 the top of their library until they reveal a land card. Put that
//	 card onto the battlefield tapped under your control and the rest
//	 into their graveyard."
//
// The end-step "if" is checked as the step begins and again as the
// ability resolves (CR 603.4), and Sméagol need not have been on the
// battlefield when the creature died (2023-06-16 ruling). With no land
// in that library, every card is revealed and all of them go to the
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b974499-bf1c-4e5b-a72c-c72220f4e591",
		Name:         "Sméagol, Helpful Guide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfACreatureDied("Sméagol, Helpful Guide — the Ring tempts you", Do(TheRingTemptsYou{})),
			Targeting(WheneverTheRingTemptsYou("Sméagol, Helpful Guide — target opponent reveals until a land", smeagolTakesALand),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// smeagolTakesALand is the second ability's body.
func smeagolTakesALand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		run, hit := revealUntil(ctx, t.ID, game.Card.IsLand, "Sméagol, Helpful Guide — revealed until a land card")
		return PutFromLibraryOntoBattlefield{
			Player: ctx.Controller(),
			Cards:  run,
			Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
			All:    true,
			Tapped: true,
			Then:   PutRestIntoGraveyard,
		}.Apply(ctx)
	}
	return nil
}
