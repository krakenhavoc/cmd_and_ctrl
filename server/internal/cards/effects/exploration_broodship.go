package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exploration_broodship_altcost is the offer a cast through Exploration
// Broodship's permission claims: the printed cost plus a land sacrificed.
const explorationBroodshipAltCost = "exploration_broodship_sacrifice_land"

// Exploration Broodship — Artifact — Spacecraft {G}, 4/4 (Edge of
// Eternities Commander):
//
//	"Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 8+.)
//	 3+ | You may play an additional land on each of your turns.
//	 8+ | Flying
//	 Once during each of your turns, you may cast a permanent spell from
//	 your graveyard by sacrificing a land in addition to paying its other
//	 costs."
//
// Station and the 8+ creature line are the shared constructors. The 3+
// extra land play is Spec.AdditionalLandPlays gated on the charge
// counters (AdditionalLandPlaysAtCharge), read live on every land play.
// The 8+ line is a gated standing graveyard permission (Serra Paragon's
// budget of one use however it is spent, PermissionTypePermanent, in your
// turn only) over nonland permanent cards, cast only. The sacrifice is
// the synthesised offer's cost (CastPermission.SacrificeLandInAddition):
// the printed mana cost and a land the caster controls, named with the
// cast, so a player with no land cannot claim it; timing, targets and the
// card's own other costs are the spell's, as the ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                    "8a83ecee-092b-456c-803f-a9aaa17c4da9",
		Name:                        "Exploration Broodship",
		Completeness:                CompletenessFull,
		Purpose:                     game.Purpose{ExtraLandDrops: 1},
		AdditionalLandPlays:         1,
		AdditionalLandPlaysAtCharge: 3,
		Activated:                   []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 4, 4),
			ThresholdKeywords(8, "flying"),
		),
		GatedCastPermissions: []game.CastPermissionGate{{
			ActiveWhen: AtChargeCounters(8),
			Permission: game.CastPermission{
				Zone:                    game.ZoneGraveyard,
				CastOnly:                true,
				Filter:                  game.PermissionFilter{NonLandPermanentOnly: true},
				PerType:                 []string{game.PermissionTypePermanent},
				Timing:                  game.TimingYourTurnOnly,
				AltCostKey:              explorationBroodshipAltCost,
				SacrificeLandInAddition: true,
				Label:                   "Once during each of your turns, cast a permanent spell from your graveyard by sacrificing a land in addition to paying its other costs (Exploration Broodship)",
			},
		}},
	})
}
