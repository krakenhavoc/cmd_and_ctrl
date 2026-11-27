package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temporal Distortion — Enchantment {3}{U}{U} (Invasion):
//
//	"Whenever a creature or land becomes tapped, put an hourglass counter on
//	 it.
//	 Each permanent with an hourglass counter on it doesn't untap during its
//	 controller's untap step.
//	 At the beginning of each player's upkeep, remove all hourglass counters
//	 from permanents that player controls."
//
// Three printed abilities, three registrations. The first is Brink of
// Disaster's "becomes tapped" watcher (a tap effect, a tap cost, or an attack
// that tapped the attacker; vigilance taps nothing) over every creature or
// land at the table. The second is Back to Basics's untap restriction over
// the counter, read as each untap step begins. The third runs at each
// player's upkeep (after their untap step) over that player's permanents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0a16783-31c2-4b2a-835d-774dc385f99a",
		Name:         "Temporal Distortion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
				AppliesTo: effects.OldB666PermanentBecameTappedCreatureOrLand,
				Key:       "Temporal Distortion — put an hourglass counter on it",
				Effect:    effects.OldB666HourglassOnTap,
			},
			effects.AtEachUpkeep("Temporal Distortion — remove all hourglass counters from permanents that player controls",
				effects.OldB666RemoveHourglassCountersOfTheActivePlayer),
		},
		UntapStepRestrictions: []game.UntapStepRestriction{
			effects.DoesntUntapDuringTheirControllersUntapSteps(effects.OldB666HasHourglassCounter()),
		},
	})
}
