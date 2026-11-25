package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Survivors' Bond — Sorcery {1}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Choose one or both —
//	 • Return target Human creature card from your graveyard to your hand.
//	 • Return target non-Human creature card from your graveyard to your
//	 hand."
//
// Unbury's graveyard bullet, once per mode (EoeABCharmReturn): each bullet
// reads its own target group (ctx.ModeTargets) and a card that left the
// graveyard in response is no longer legal (CR 608.2b), so only that bullet
// does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff53c3a0-988a-40a4-b5f9-e5a4b634f5eb",
		Name:         "Survivors' Bond",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target Human creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target Human creature card from your graveyard", effects.YouOwn(), effects.Creature(), effects.HasSubtype("Human")), effects.EoeABCharmReturn),
			effects.ModeDoing("Return target non-Human creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target non-Human creature card from your graveyard", effects.YouOwn(), effects.Creature(), effects.Not(effects.HasSubtype("Human"))), effects.EoeABCharmReturn),
		),
	})
}
