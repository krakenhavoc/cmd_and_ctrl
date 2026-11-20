package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of Fell Rites — Creature — Human Warlock {W}{B}, 2/2:
//
//	"{T}, Pay 3 life, Sacrifice this creature: Return target creature
//	 card from your graveyard to the battlefield. Activate only as a
//	 sorcery.
//	 Unearth {3}{W}{B} ({3}{W}{B}: Return this card from your graveyard
//	 to the battlefield. It gains haste. Exile it at the beginning of
//	 the next end step or if it would leave the battlefield. Unearth
//	 only as a sorcery.)"
//
// A one-shot reanimator body, and a reanimator body that can reanimate
// ITSELF a second time (unearth it, sacrifice it to bring back
// something bigger, and it exiles at end of turn regardless). The
// activated ability's cost is Plus(TapCost(), PayLife(3),
// SacrificeThis()) — all three validated before any of them are paid,
// so a Priest that cannot afford the life is never tapped for nothing
// (CR 602.5a). Unearth is the constructor: the return, the haste, the
// end-step exile and the leaves-the-battlefield redirect all come with
// the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec6ccd8f-cc75-4a50-8639-f4ec31a280aa",
		Name:         "Priest of Fell Rites",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{T}, Pay 3 life, Sacrifice this creature: Return target creature card from your graveyard to the battlefield.",
				Cost:         effects.Plus(effects.TapCost(), effects.PayLife(3), effects.SacrificeThis()),
				SorcerySpeed: true,
				Targets:      effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				Effect:       effects.ReturnFirstLegalGraveyardTargetToBattlefield,
			},
			effects.Unearth("{3}{W}{B}"),
		},
	})
}
