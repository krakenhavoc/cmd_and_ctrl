package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// On Thin Ice — Snow Enchantment — Aura {W} (Modern Horizons):
//
//	"Enchant snow land you control
//	 When this Aura enters, exile target creature an opponent controls until
//	 this Aura leaves the battlefield."
//
// Hardlight Containment's shape. "Until this Aura leaves the battlefield" is
// CR 610.3: the exile is a one-shot and the return a second one-shot created
// as the Aura leaves, so the permanent is back before anyone has priority; it
// returns under its owner's control (CR 610.3c), and an Aura removed before its
// trigger resolves exiles nothing (CR 610.3b). The enchant clause is a real
// restriction re-checked as a state-based action (CR 303.4c): a land that stops
// being snow, or stops being yours, sends the Aura to the graveyard, which
// returns the creature. Being snow, the Aura itself is a snow permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6736bb96-6718-48bb-9385-789d3349dc20",
		Name:         "On Thin Ice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant snow land you control", effects.Land(), effects.Snow(), effects.YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				Key:       "On Thin Ice — exile target creature an opponent controls until this Aura leaves the battlefield",
				Effect: effects.ExileChosenTargetUntilThisLeaves(
					"On Thin Ice — the exiled card returns when On Thin Ice leaves the battlefield"),
			},
		},
	})
}
