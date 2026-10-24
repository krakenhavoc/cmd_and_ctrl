package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathgreeter — Creature — Human Shaman {B}, 1/1:
//
//	"Whenever another creature dies, you may gain 1 life."
//
// Any other creature, anyone's, token or card (AnotherCreatureDied).
// The "you may" is a choice made when the ability RESOLVES (CR 608.2),
// so the trigger always goes on the stack and the question is a
// MayChoice inside it. It is deliberately not Optional(...), the early
// prompt that asks before the ability is put on the stack (#2494).
// Each death is its own trigger, so a wrath is one question per
// creature; a creature that was exiled instead of dying never reaches
// the graveyard-only EventLTB test.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7ad11dc-c962-4034-b15a-a080c6fff75f",
		Name:         "Deathgreeter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AnotherCreatureDied, "Deathgreeter — you may gain 1 life", deathgreeterMayGain),
		},
	})
}

func deathgreeterMayGain(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Deathgreeter — gain 1 life?",
		Then:     mayGainLife,
		Word:     "1",
	}.Apply(NewContext(g, item))
}
