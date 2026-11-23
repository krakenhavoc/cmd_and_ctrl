package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Misery's Shadow — Creature — Shade {1}{B}, 2/2 (The Brothers' War):
//
//	"If a creature an opponent controls would die, exile it instead.
//	 {1}: This creature gets +1/+1 until end of turn."
//
// The replacement is Head of the Hunt's (rfOpponentCreatureWouldDie): a
// creature that would go from the battlefield to the graveyard goes to
// exile, never having died (CR 700.4), tokens included. The pump is the
// shared self-pump row, activatable any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "928c62fe-9c4d-4e89-bd7a-0d3b6a81f393",
		Name:         "Misery's Shadow",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.RfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Misery's Shadow: exile instead of dying",
		}},
		Activated: []effects.ActivatedAbility{effects.LegSelfPump("{1}", 1, "This creature", "Misery's Shadow")},
	})
}
