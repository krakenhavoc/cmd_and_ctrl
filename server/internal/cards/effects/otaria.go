package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otaria — Plane — Dominaria (Planechase Anthology, #4870):
//
//	"Instant and sorcery cards in graveyards have flashback. The flashback
//	 cost is equal to the card's mana cost.
//	 Whenever chaos ensues, take an extra turn after this one."
//
// A plane (ADR 0216 §6). The static is Lier's standing flashback permission
// (instants and sorceries only, the card's own mana cost, exiled whenever it
// leaves the stack), held by EVERY seat over its own graveyard as long as
// the plane is face up in the command zone (the planar walk of
// standingCastPermissionsLocked, CR 901.7). The chaos ability gives the
// planar controller an extra turn per chaos rolled (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29e3371d-aa41-46a9-921a-96565be47eff",
		Name:         "Otaria",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:              game.ZoneGraveyard,
			Filter:            game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true},
			AltCostKey:        "flashback",
			ExileOnResolution: true,
			Label:             "Flashback — its mana cost (Otaria)",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Otaria — take an extra turn after this one", Do(TakeExtraTurn{})),
		},
	})
}
