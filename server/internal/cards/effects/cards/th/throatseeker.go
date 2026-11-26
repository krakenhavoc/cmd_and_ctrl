package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Throatseeker — Creature — Vampire Ninja {2}{B}, 3/2 (Modern Horizons):
//
//	"Unblocked attacking Ninjas you control have lifelink."
//
// A layer 6 keyword grant over the Ninjas (CR 205.3m, read through the layers)
// its controller controls that are attacking and unblocked (CR 509.1h, 509.3g):
// an attacker whose defending player has finished declaring blockers and that
// nothing blocks. Throatseeker is a Ninja itself. The grant reads combat, so the
// static declares that it depends on attacking status and on blocks being
// declared; a blocked attacker has no lifelink, and a Ninja that is not
// attacking has none either.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.HasSubtype(source.TextSubtype("Ninja")) && target.Controller == source.Controller &&
			target.AttackingTarget != uuid.Nil && g.UnblockedAttackerForEffect(target.InstanceID)
	}, "lifelink")
	grant.Label = "Unblocked attacking Ninjas you control have lifelink."
	grant.DependsOnAttackingStatus = true
	grant.DependsOnEvents = []game.EventKind{game.EventBlock, game.EventStepBegan}
	effects.Register(effects.Spec{
		OracleID:     "edb7bd88-971a-4fcc-9d76-d73f21a4926a",
		Name:         "Throatseeker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
