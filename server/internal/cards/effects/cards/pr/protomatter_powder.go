package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protomatter Powder — {2}{U} Artifact (Shards of Alara, S82 OLD b516):
//
//	"{4}{W}, {T}, Sacrifice this artifact: Return target artifact card from
//	 your graveyard to the battlefield."
//
// The target is chosen as the ability is activated and re-checked as it
// resolves (CR 608.2b); the Powder is sacrificed as a cost, so it can't return
// itself (it is not in the graveyard yet when the target is chosen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22d57bc8-e473-47c5-8e48-c0fb98e66948",
		Name:         "Protomatter Powder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}, {T}, Sacrifice this artifact: Return target artifact card from your graveyard to the battlefield.",
			Cost:    effects.Plus(effects.ManaCost("{4}{W}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect:  effects.OldB516ReturnTargetArtifactToBattlefield,
		}},
	})
}
