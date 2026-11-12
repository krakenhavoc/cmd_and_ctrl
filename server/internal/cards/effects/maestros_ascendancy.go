package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// maestrosAscendancyAltCost is the key the sacrifice offer is claimed under.
// An on-disk identity: never renamed, never reused.
const maestrosAscendancyAltCost = "maestros_ascendancy_sacrifice_creature"

// Maestros Ascendancy — Enchantment {U}{B}{R} (Streets of New Capenna):
//
//	"Once during each of your turns, you may cast an instant or sorcery
//	 spell from your graveyard by sacrificing a creature in addition to
//	 paying its other costs. If a spell cast this way would be put into
//	 your graveyard, exile it instead."
//
// A standing graveyard permission (ADR 0066) with Johann's one use per turn
// per Ascendancy object (PerType PermissionTypePermanent, which asks no
// question), open in your own turn only, over the cards whose spell is an
// instant or a sorcery. The sacrifice is the synthesised offer's cost
// (CastPermission.SacrificeCreatureInAddition, #5670): the printed mana cost
// and a creature the caster controls, sacrificed as a cost (CR 118.8) and
// named when the cast is announced, so a player with no creature cannot
// claim it (CR 118.3) and a payoff sees the sacrifice. Timing, targets and
// the card's own additional costs are the spell's. The last sentence is
// Torrential Gearhulk's replacement (ExileInsteadOfGraveyard): however the
// spell would reach your graveyard (it resolves, is countered or fizzles),
// it is exiled instead; a spell returned to a hand goes there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1438681-241b-4d53-9470-4d04a7797ea8",
		Name:         "Maestros Ascendancy",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                        game.ZoneGraveyard,
			CastOnly:                    true,
			Filter:                      game.PermissionFilter{InstantOrSorceryOnly: true},
			PerType:                     []string{game.PermissionTypePermanent},
			Timing:                      game.TimingYourTurnOnly,
			AltCostKey:                  maestrosAscendancyAltCost,
			SacrificeCreatureInAddition: true,
			ExileInsteadOfGraveyard:     true,
			Label:                       "Once during each of your turns, cast an instant or sorcery spell from your graveyard by sacrificing a creature in addition to paying its other costs (Maestros Ascendancy)",
		}},
	})
}
