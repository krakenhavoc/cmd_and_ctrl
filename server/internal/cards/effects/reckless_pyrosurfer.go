package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "65c35934-796c-4495-a6b5-eea654dfcd55",
		Name:            "Reckless Pyrosurfer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []AbilityGrant{{
			Key:       recklessPyrosurferBattleCry,
			Triggered: []game.TriggeredAbility{BattleCry("Reckless Pyrosurfer")},
			Text:      "Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)",
		}},
		Triggered: []game.TriggeredAbility{
			Landfall("Reckless Pyrosurfer — this creature gains battle cry until end of turn",
				thisStillHere(func(g *game.Game, item *game.StackItem) error {
					return GrantAbilitiesFor{
						Target: item.SourceCardID,
						Keys:   []string{recklessPyrosurferBattleCry},
						Label:  "Reckless Pyrosurfer — battle cry until end of turn",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
