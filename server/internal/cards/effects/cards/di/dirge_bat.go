package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dirge Bat — Creature — Bat {2}{B}{B}, 3/3 (Ikoria: Lair of Behemoths,
// #5512):
//
//	"Mutate {4}{B}{B} (If you cast this spell for its mutate cost, put it over
//	 or under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Flash
//	 Flying
//	 Whenever this creature mutates, destroy target creature or planeswalker an
//	 opponent controls."
//
// Mutate is the engine's (ADR 0223), and Gemrazer's trigger shape over a
// creature or planeswalker an opponent controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4746828f-46ea-439f-88de-5e54cc536545",
		Name:             "Dirge Bat",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash", "flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{4}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Dirge Bat — destroy target creature or planeswalker an opponent controls", effects.DestroyChosenPermanent),
				effects.Harmful(effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())))),
		},
	})
}
