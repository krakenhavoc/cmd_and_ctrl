package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade Rallier — Creature — Human Warrior {1}{G}{W}, 3/2:
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, return target permanent
//	 card with mana value 2 or less from your graveyard to the
//	 battlefield."
//
// Revolt is an intervening if (revolt.go). The return is not optional.
// The card returns under its owner's control, which is you: the target
// is in your graveyard.
//
// No simplification.
func init() {
	ability := effects.WhenThisEntersIfRevolt("Renegade Rallier — return a permanent card from your graveyard to the battlefield",
		// The same "return the legal target to the battlefield" body Sheoldred uses.
		effects.SheoldredReanimate)
	ability.Targets = effects.TargetCardInGraveyard(
		"target permanent card with mana value 2 or less in your graveyard",
		effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(2),
	)
	effects.Register(effects.Spec{
		OracleID:     "6fa07b6c-f01a-4416-b0fc-986b0fc4e412",
		Name:         "Renegade Rallier",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{ability},
	})
}
