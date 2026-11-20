package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arbalest Elite — Creature — Human Archer {2}{W}{W}, 2/3 (Magic 2012,
// #5573):
//
//	"{2}{W}, {T}: This creature deals 3 damage to target attacking or
//	 blocking creature. This creature doesn't untap during your next
//	 untap step."
//
// Elite Archers' damage row, then the source sits out its controller's
// next untap step (DoesntUntapNextUntapStep, Spectral Bears' freeze). The
// freeze is on the Elite itself, so it applies even when the target has
// left and only the damage is skipped; an Elite that left the battlefield
// in response is a new object and is not frozen (#1432).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d79b7a8-2b4c-4893-afdb-879c74957288",
		Name:         "Arbalest Elite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: This creature deals 3 damage to target attacking or blocking creature. This creature doesn't untap during your next untap step.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  arbalestEliteShoots,
		}},
	})
}

// arbalestEliteShoots deals the 3 damage to the target still legal, then
// keeps the Elite tapped through its controller's next untap step.
func arbalestEliteShoots(g *game.Game, item *game.StackItem) error {
	if err := sourceDealsDamageToEachLegalTarget(3)(g, item); err != nil {
		return err
	}
	return DoesntUntapNextUntapStep{
		Targets: []uuid.UUID{item.SourceCardID},
		Player:  item.Controller,
		Label:   "Arbalest Elite — doesn't untap during your next untap step",
	}.Apply(NewContext(g, item))
}
