package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grifter's Blade — Artifact — Equipment {3} (Ravnica: City of Guilds, S82 OLD
// b582):
//
//	"Flash
//	 As this Equipment enters, choose a creature you control it could be
//	 attached to. If you do, it enters attached to that creature.
//	 Equipped creature gets +1/+1.
//	 Equip {1}"
//
// CR 614.1c makes "as this enters" a replacement effect, so the choice is the
// EntryCardChoose of ADR 0170: the prompt opens while the entry is paused and
// the named creature rides the entry (ReplacementEvent.AttachAtETB, the host
// an Aura put onto the battlefield carries), so the Blade is attached before
// anything sees it arrive (CR 614.12). The candidates are the creatures its
// controller controls that it could be attached to, so one with protection
// from artifacts (or colorless) is not offered (CR 702.16d); with none, the
// Blade enters unattached.
//
// No simplification.
func init() {
	const name = "Grifter's Blade"
	probe := game.Card{Name: name, TypeLine: "Artifact — Equipment", ManaCost: "{3}"}
	effects.Register(effects.Spec{
		OracleID:        "fe8b7835-a9b2-4a8d-964a-471430f1d143",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.AsEntersChoosePermanents(name,
				"Grifter's Blade — as it enters, choose a creature you control it could be attached to. If you do, it enters attached to that creature.",
				&game.EntryCardChoice{
					Action: game.EntryCardChoose,
					Matches: func(c game.Card) bool {
						return c.IsCreature() && !game.ProtectedFromAttachment(&c, &probe)
					},
					Min:  1,
					Max:  1,
					Then: grifterBladeAttach,
				}),
		},
		Static:    []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

// grifterBladeAttach is the entry's Then: the Blade arrives attached to the
// creature that was named.
func grifterBladeAttach(_ *game.Game, ev *game.ReplacementEvent, picked []uuid.UUID) error {
	if len(picked) > 0 {
		ev.AttachAtETB(picked[0])
	}
	return nil
}
