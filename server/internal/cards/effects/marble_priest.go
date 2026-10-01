package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marble Priest — Artifact Creature — Cleric, {5}, 3/3:
//
//	"All Walls able to block this creature do so.
//	 Prevent all combat damage that would be dealt to this creature by
//	 Walls."
//
// #1684: the filtered Lure — FilteredLure(game.BlockerFilterWall), a
// registered filter KEY on the requirement, so only the defender's
// Walls are made to block it and any other creature is free. The
// second line is a standing CR 615 prevention (Caduceus's shape),
// narrowed to combat damage whose source is a Wall when it is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96bcabee-e84e-409f-8439-80f5308d73e9",
		Name:         "Marble Priest",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{FilteredLure(game.BlockerFilterWall)},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageTarget != src.InstanceID {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.DamageSource)
				return ok && dealer.HasSubtype("Wall")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Marble Priest — prevent combat damage dealt to it by Walls",
		}},
	})
}
