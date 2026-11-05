package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Badgey, Malicious Glitch — Legendary Artifact Creature — Illusion Horror
// {3}{B}, 4/3 (Star Trek Commander, #4974):
//
//	"Whenever Badgey becomes tapped, each opponent loses 1 life and you
//	 gain 1 life."
//
// "Becomes tapped" is any tap, an attack included (ThisBecameTapped,
// CR 701.26a); an attack that does not tap it (vigilance) is silent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "227fb13b-0a07-4937-9048-7cf02efbc8e3",
		Name:         "Badgey, Malicious Glitch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Badgey, Malicious Glitch — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
