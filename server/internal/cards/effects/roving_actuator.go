package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roving Actuator — Artifact Creature — Robot {3}{R}, 3/4 (Edge of
// Eternities):
//
//	"Void — When this creature enters, if a nonland permanent left the
//	 battlefield this turn or a spell was warped this turn, exile up to
//	 one target instant or sorcery card with mana value 2 or less from
//	 your graveyard. Copy it. You may cast the copy without paying its
//	 mana cost."
//
// Void is an intervening if (void.go). "Up to one target" may be none, in
// which case the ability does nothing. The exile goes through the CR 614
// window and only a card that really reached exile is copied
// (CopyAndCast, ADR 0066's 2026-10-13 card-copy amendment); the copy is
// offered for free and is cast, or not, while the ability is still
// resolving (CR 707.12, CR 608.2g). A copy not cast ceases to exist and the
// original stays in exile.
//
// No simplification.
func init() {
	enters := WhenThisEntersIfVoid("Roving Actuator — exile up to one instant or sorcery card from your graveyard, copy it, and you may cast the copy",
		func(g *game.Game, item *game.StackItem) error {
			ids := legalTargetIDs(NewContext(g, item))
			if len(ids) == 0 {
				return nil
			}
			return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
				return CopyAndCast{Cards: exiled, Free: true}.Apply(NewContext(g, item))
			})
		})
	enters.Targets = TargetCardInGraveyard("up to one target instant or sorcery card with mana value 2 or less in your graveyard",
		YouOwn(), Or(Instant(), Sorcery()), ManaValueLE(2)).WithCount(0, 1)
	Register(Spec{
		OracleID:     "78ac1e6c-2704-4844-9e87-642324c23997",
		Name:         "Roving Actuator",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
