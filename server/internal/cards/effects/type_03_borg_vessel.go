package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	cast := WheneverYouCast(nil, "Type-03 Borg Vessel — create a 1/1 colorless Borg artifact creature token",
		Do(CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: 1}))
	cast.ActiveWhen = AtChargeCounters(7)
	Register(Spec{
		OracleID:     "ff2ab17e-185b-4782-a4e2-e8ace40ef1bb",
		Name:         "Type-03 Borg Vessel",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact spells you cast cost {1} less to cast.", YourSpell(), ArtifactSpell()),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 5, 5),
			ThresholdKeywords(7, "flying"),
		),
		Triggered: []game.TriggeredAbility{cast},
	})
}
