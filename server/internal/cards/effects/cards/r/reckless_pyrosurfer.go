package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const recklessPyrosurferBattleCry = "reckless-pyrosurfer/battle-cry"

// Reckless Pyrosurfer — Creature — Human Scout {1}{R}, 2/2 (Modern
// Horizons 3):
//
//	"Haste
//	 Landfall — Whenever a land you control enters, this creature gains
//	 battle cry until end of turn. (Whenever this creature attacks, each
//	 other attacking creature gets +1/+0 until end of turn. Each instance
//	 of battle cry triggers separately.)"
//
// Battle cry is the keyword's trigger (BattleCry), declared as an ADR
// 0093 bundle and granted for the turn to this Pyrosurfer as the landfall
// trigger resolves. Each landfall grants another instance, and each
// instance triggers separately (CR 702.91b); the grant goes to the
// permanent that triggered, so a Pyrosurfer that has left (CR 400.7) is
// not given anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65c35934-796c-4495-a6b5-eea654dfcd55",
		Name:            "Reckless Pyrosurfer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []effects.AbilityGrant{{
			Key:       recklessPyrosurferBattleCry,
			Triggered: []game.TriggeredAbility{effects.BattleCry("Reckless Pyrosurfer")},
			Text:      "Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Reckless Pyrosurfer — this creature gains battle cry until end of turn",
				effects.ThisStillHere(func(g *game.Game, item *game.StackItem) error {
					return effects.GrantAbilitiesFor{
						Target: item.SourceCardID,
						Keys:   []string{recklessPyrosurferBattleCry},
						Label:  "Reckless Pyrosurfer — battle cry until end of turn",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
