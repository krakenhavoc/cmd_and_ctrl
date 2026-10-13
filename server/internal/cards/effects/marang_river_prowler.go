package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marang River Prowler — Creature — Human Rogue {2}{U}, 2/1:
//
//	"This creature can't block and can't be blocked.
//	 You may cast this card from your graveyard as long as you control
//	 a black or green permanent."
//
// A Spec.CastableZonesIf card (2026-10-13 amendment to ADR 0066). The
// colour test reads the permanent's effective colours (layer 5), so an
// artifact turned green counts and a face-down permanent, which is
// colourless, does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "002e767c-0b30-4ad1-a924-88da400a4cff",
		Name:         "Marang River Prowler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
			RestrictSelf(game.CantBeBlocked),
		},
		CastableZonesIf: []game.ConditionalCastZone{{Zone: game.ZoneGraveyard, Cond: ownerControlsABlackOrGreenPermanent}},
	})
}

// ownerControlsABlackOrGreenPermanent is "as long as you control a
// black or green permanent", asked of the card's owner.
func ownerControlsABlackOrGreenPermanent(g *game.Game, owner uuid.UUID, _ game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && (c.HasColor("B") || c.HasColor("G")) {
			return true
		}
	}
	return false
}
