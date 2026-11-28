package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yawgmoth's Agenda — Enchantment {3}{B}{B} (Invasion):
//
//	"You can't cast more than one spell each turn.
//	 You may play lands and cast spells from your graveyard.
//	 If a card would be put into your graveyard from anywhere, exile it
//	 instead."
//
// Moderation's one-spell cast restriction (counted off the turn's cast
// tally, every zone), Colfenor's Plans's standing cast permission over the
// graveyard with lands as well as spells, and Festival of Embers's graveyard
// replacement for the controller's graveyard alone (CR 614.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72177c93-ed1f-47bd-99a6-ad229a292d46",
		Name:         "Yawgmoth's Agenda",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Yawgmoth's Agenda — you can't cast more than one spell each turn.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && q.Game.CastTallyFor(q.Controller).Total >= 1
			},
		}},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: game.CastPermission{
				Zone:  game.ZoneGraveyard,
				Label: "You may play lands and cast spells from your graveyard (Yawgmoth's Agenda)",
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.GraveyardBecomesExile{
				YoursOnly: true,
				Label:     "Yawgmoth's Agenda: exile instead of your graveyard",
			}.Build(),
		},
	})
}
