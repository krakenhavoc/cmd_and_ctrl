package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindful Biomancer — Creature — Dryad Druid {1}{G}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, you gain 1 life.
//	 {2}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// The pump is a layer-7c scoped effect on the Biomancer. "Once each turn" is
// the per-object activation tally (OncePerTurnActivation), read at the
// announce, so a second activation in the same turn is refused whoever has
// priority; the tally is per permanent, so a new object (it left and came
// back) can activate again.
//
// No simplification.
func init() {
	const label = "{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "03af9d24-2df6-4318-82f2-8abdd22206f1",
		Name:         "Mindful Biomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mindful Biomancer — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:      ManaCost("{2}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Mindful Biomancer — +2/+2 until end of turn"),
		}},
	})
}
