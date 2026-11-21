package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Type-03 Borg Vessel — Artifact — Spacecraft {3}{W}, 5/5 (Star Trek
// Commander, #4974):
//
//	"Artifact spells you cast cost {1} less to cast.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 7+.)
//	 7+ | Flying
//	 Whenever you cast a spell, create a 1/1 colorless Borg artifact
//	 creature token."
//
// The reduction is Foundry Inspector's, above the bar, so it works at any
// counter count. The station body is The Seriema's, and the cast trigger
// is part of the striated 7+ block (CR 721.2a): it exists only from seven
// counters on, and then fires for every spell its controller casts.
//
// No simplification.
func init() {
	cast := effects.WheneverYouCast(nil, "Type-03 Borg Vessel — create a 1/1 colorless Borg artifact creature token",
		effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Borg artifact"), N: 1}))
	cast.ActiveWhen = effects.AtChargeCounters(7)
	effects.Register(effects.Spec{
		OracleID:     "ff2ab17e-185b-4782-a4e2-e8ace40ef1bb",
		Name:         "Type-03 Borg Vessel",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(7, 5, 5),
			effects.ThresholdKeywords(7, "flying"),
		),
		Triggered: []game.TriggeredAbility{cast},
	})
}
