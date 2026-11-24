package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf the White — Legendary Creature — Avatar Wizard {3}{W}{W}, 4/5:
//
//	"Flash
//	 You may cast legendary spells and artifact spells as though they had
//	 flash.
//	 If a legendary permanent or an artifact entering or leaving the
//	 battlefield causes a triggered ability of a permanent you control to
//	 trigger, that ability triggers an additional time."
//
// Flash is PrintedKeywords. The second ability is two cast-timing rules,
// one per kind of spell. The third is a CR 603.2d doubler over both kinds
// of event: the entering object is judged as it is on the battlefield,
// the leaving one as it last existed there (CR 603.10a).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b3b5f6a-375e-4479-9384-a942eda83f9b",
		Name:            "Gandalf the White",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{LegendaryOnly: true},
				"You may cast legendary spells as though they had flash."),
			effects.CastKindAsThoughFlash(game.PermissionFilter{CardType: "artifact"},
				"You may cast artifact spells as though they had flash."),
		},
		TriggerDoublers: []game.TriggerDoubler{gandalfTheWhiteDoubler()},
	})
}

func gandalfTheWhiteDoubler() game.TriggerDoubler {
	d := effects.DoublesEnteringOrLeaving(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLegendary() || c.IsArtifact()
	}, "Abilities that trigger when a legendary permanent or an artifact enters or leaves the battlefield trigger an additional time.")
	d.Label = "Gandalf the White"
	return d
}
