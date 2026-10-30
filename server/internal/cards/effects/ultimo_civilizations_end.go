package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultimo, Civilization's End — Legendary Artifact Creature — Robot Villain
// {4}{B}{B}, 6/5 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 When Ultimo enters, each opponent sacrifices a creature of their choice.
//	 {2}{B}, Discard this card: Each opponent sacrifices a creature of their
//	 choice."
//
// Both halves are the same untargeted edict (EachPlayerSacrifices over every
// player but the controller, each choosing their own). The second is a hand
// ability (Sureshot Sower's shape: it is activated from the hand and the
// card itself is the discard cost), so it works at instant speed, and the
// card is gone from the hand whether or not anything can be sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43d79ee6-8ba1-4ec9-9bb2-907ed8401a2f",
		Name:            "Ultimo, Civilization's End",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Ultimo — each opponent sacrifices a creature", eachOpponentSacrificesACreature)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Discard this card: Each opponent sacrifices a creature of their choice.",
			Cost:    game.AbilityCost{Mana: "{2}{B}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  eachOpponentSacrificesACreature,
		}},
	})
}

// eachOpponentSacrificesACreature is "each opponent sacrifices a creature
// of their choice": every other player chooses and sacrifices their own.
func eachOpponentSacrificesACreature(g *game.Game, item *game.StackItem) error {
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            Creature(),
		Label:            "a creature",
	}.Apply(NewContext(g, item))
}
