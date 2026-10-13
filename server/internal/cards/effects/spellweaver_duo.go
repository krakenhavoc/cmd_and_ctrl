package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellweaver Duo — Creature — Human Wizard {6}{U}:
//
//	"Assist (Another player can pay up to {6} of this spell's cost.)
//	 When this creature enters, you may return target tapped creature to
//	 its owner's hand."
//
// Angel of Serenity's optional targeted ETB: the "you may" is the
// trigger's yes/no prompt (CR 603.5), and the pick is a tapped creature
// on the battlefield. The bounce goes to the owner's hand.
//
// The tapped test is read off the card's Tapped flag on the candidate
// (a clause of this card alone, kept beside it as b41Attacking is).
func init() {
	Register(Spec{
		OracleID:        "1fd4204f-f627-4919-9d48-79b172f4560e",
		Name:            "Spellweaver Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Spellweaver Duo — return target tapped creature to its owner's hand?"},
			Targets:        TargetCreature("target tapped creature", spellweaverDuoTapped),
			Key:            "Spellweaver Duo — return target tapped creature to its owner's hand",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return bounceTheTarget(item, NewContext(g, item))
			},
		}},
	})
}

// spellweaverDuoTapped is "tapped": the candidate's Tapped flag, read
// as the pick is judged at announce and again at resolution.
func spellweaverDuoTapped(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.Tapped
}
