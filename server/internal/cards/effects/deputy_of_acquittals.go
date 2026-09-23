package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deputy of Acquittals — "Flash. When this creature enters, you may
// return another target creature you control to its owner's hand."
//
// "ANOTHER" is exact. The clause is built per trigger through
// AnotherTarget (TriggeredAbility.TargetsFrom, which is handed the
// source), so the picker excludes THIS Deputy by instance rather than
// by name: a second Deputy — a Clone or a token copy of this one — is
// a legal target, as printed, and the Deputy whose trigger it is
// never is. The resolution re-check (CR 608.2b) runs the same clause.
//
// The "you may" is the optional prompt, asked before the target is
// chosen (CR 603.3c-d): no, and nothing happens.
func init() {
	Register(Spec{
		OracleID:        "3cbb5045-8566-4279-b7d3-3e599b11ccc5",
		Name:            "Deputy of Acquittals",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
				return TargetCreature("another target creature you control", YouControl(), other)
			}),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Deputy of Acquittals — return a creature to hand",
					bounceChosenTarget)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Deputy of Acquittals — return another creature you control to its owner's hand?",
			},
		}},
	})
}
