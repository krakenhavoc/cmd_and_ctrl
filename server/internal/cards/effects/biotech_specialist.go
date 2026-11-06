package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biotech Specialist — Creature — Insect Scientist {R}{G}, 1/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, create a Lander token. (It's an
//	 artifact with "{2}, {T}, Sacrifice this token: Search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle.")
//	 Whenever you sacrifice an artifact, this creature deals 2 damage
//	 to target opponent."
//
// The Lander is Horizon Explorer's (b16LanderToken), so the card feeds
// its own second ability. That one is Crime Novelist's "whenever you
// sacrifice an artifact" (b12YouSacrificedAnArtifact: one trigger per
// artifact sacrificed, whatever asked for the sacrifice) with Gastal
// Thrillseeker's target opponent, and the Specialist is the damage's
// source. With no opponent to target the trigger is removed
// (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78de77ef-c325-4342-869f-042921229577",
		Name:         "Biotech Specialist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Biotech Specialist — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
			Targeting(On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12YouSacrificedAnArtifact(ev, source, g)
			}, "Biotech Specialist — 2 damage to target opponent", eoeABTwoDamageToTargetOpponent),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// eoeABTwoDamageToTargetOpponent is "this creature deals 2 damage to
// target opponent".
func eoeABTwoDamageToTargetOpponent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: victim, Amount: 2}.Apply(ctx)
}
