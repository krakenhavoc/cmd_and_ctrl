package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Junk Jet — Artifact — Equipment {1}{R} (Fallout, #5948):
//
//	"When this Equipment enters, create a Junk token.
//	 {3}, Sacrifice another artifact: Double equipped creature's power
//	 until end of turn.
//	 Equip {1}"
//
// The Junk is the catalog's (Break Down). The activated ability reads the
// equipped creature's power as it resolves and adds that much until end of
// turn (Bulk Up's doubling); with nothing equipped by then it does nothing.
// The sacrifice is a cost, so the artifact is gone whatever happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0347b8b7-0553-4a1f-9677-ad730470c345",
		Name:         "Junk Jet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Junk Jet — create a Junk token",
				effects.Do(effects.CreateToken{Template: effects.JunkToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, Sacrifice another artifact: Double equipped creature's power until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.SacrificeAnotherN(1, "another artifact", effects.Artifact())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					host := effects.AttachedHostFor(g, item.SourceCardID)
					if host == nil || !host.IsCreature() {
						return nil
					}
					return effects.BoostUntilEOT{Target: host.InstanceID, Power: host.CurrentPower(), Label: "Junk Jet — double power"}.Apply(effects.NewContext(g, item))
				},
			},
			effects.EquipAbility("{1}"),
		},
	})
}
