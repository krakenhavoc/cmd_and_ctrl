package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karador, Ghost Chieftain — Legendary Creature — Centaur Spirit
// {5}{W}{B}{G}, 3/4 (Commander Masters):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 Once during each of your turns, you may cast a creature spell from
//	 your graveyard."
//
// The cost reduction is Ghoultree's, counted as the spell is cast (CR 601.2f).
// The permission is Gisa and Geralf's with no creature type: one creature
// spell from your graveyard on each of your turns, at sorcery timing like
// any creature spell, paid for in full (it is not an alternative cost). The
// permission is read from the battlefield, so it ends when Karador leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "304c194f-fa0b-42e5-93ca-8c587181d33c",
		Name:         "Karador, Ghost Chieftain",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CreatureCardsInYourGraveyard,
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{CreatureOnly: true},
			PerType:  []string{game.PermissionTypeCreature},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Once during each of your turns, cast a creature spell from your graveyard (Karador, Ghost Chieftain)",
		}},
	})
}
