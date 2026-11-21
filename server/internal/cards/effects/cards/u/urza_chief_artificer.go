package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza, Chief Artificer — Legendary Creature — Human Artificer
// {3}{W}{U}{B} (The Brothers' War Commander, #5482):
//
//	"Affinity for artifact creatures
//	 Artifact creatures you control have menace.
//	 At the beginning of your end step, create a 0/0 colorless Construct
//	 artifact creature token with 'This token gets +1/+1 for each
//	 artifact you control.'"
//
// Affinity is Frogmite's self cost reduction, counting artifact creatures
// (CR 702.41a). The menace grant reads types through the layered view, so
// Urza himself (not an artifact) is not granted it and a Construct is. The
// Construct is Urza, Lord High Artificer's token (UrzaConstructToken).
//
// No simplification.
func init() {
	menace := effects.KeywordGrant(effects.TrekB34ArtifactCreaturesYouControl, "menace")
	menace.Label = "Artifact creatures you control have menace."
	effects.Register(effects.Spec{
		OracleID:     "9deced63-80d0-419e-a5ac-71b06ba1459e",
		Name:         "Urza, Chief Artificer",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifact creatures", effects.And(effects.Artifact(), effects.Creature())),
		},
		Static: []game.StaticAbility{menace},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Urza, Chief Artificer — create a Construct",
				effects.Do(effects.CreateToken{Template: effects.UrzaConstructToken(), N: 1})),
		},
	})
}
