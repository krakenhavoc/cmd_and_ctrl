package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Festival of Embers — Enchantment {4}{R} (Bloomburrow):
//
//	"During your turn, you may cast instant and sorcery spells from your
//	 graveyard by paying 1 life in addition to their other costs.
//	 If a card or token would be put into your graveyard from anywhere, exile it
//	 instead.
//	 {1}{R}: Sacrifice this enchantment."
//
// The first line is Noctis's standing graveyard permission (CastPermission
// with LifeInAddition) filtered to instants and sorceries and gated on it being
// your turn (CastPermissionGate.Condition), so the cost is the spell's own mana
// cost plus 1 life (a player at 0 life cannot, CR 119.4) and every other rule of
// the spell's timing stands. The second is Rest in Peace for your graveyard only
// (GraveyardBecomesExile YoursOnly, cards and tokens alike, CR 614.1a): a creature
// of yours that would die is exiled and never died, and a spell cast from the
// graveyard is exiled when it would return there. The third is Hopeful Vigil's
// sacrifice, which Festival pays on its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2de138b0-3159-4a0c-a474-a4e5600b2e51",
		Name:         "Festival of Embers",
		Completeness: CompletenessFull,
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: game.CastPermission{
				Zone:           game.ZoneGraveyard,
				CastOnly:       true,
				Filter:         game.PermissionFilter{InstantOrSorceryOnly: true},
				AltCostKey:     "festival_of_embers",
				LifeInAddition: 1,
				Label:          "Cast an instant or sorcery spell from your graveyard, paying 1 life in addition (Festival of Embers)",
			},
			Condition: func(g *game.Game, you uuid.UUID, _ uuid.UUID) bool {
				return g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) && g.Seats[g.Turn.ActiveSeat].ID == you
			},
		}},
		Replacements: []game.ReplacementEffect{
			GraveyardBecomesExile{
				YoursOnly: true,
				Label:     "Festival of Embers: exile instead of your graveyard",
			}.Build(),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: Sacrifice this enchantment.",
			Cost:    ManaCost("{1}{R}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  sacrificeDelayedTriggerSource,
		}},
	})
}
