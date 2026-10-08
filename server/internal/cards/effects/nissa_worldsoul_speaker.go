package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nissa, Worldsoul Speaker — Legendary Creature — Elf Druid {3}{G}, 3/3:
//
//	"Landfall — Whenever a land you control enters, you get {E}{E} (two
//	 energy counters).
//	 You may pay eight {E} rather than pay the mana cost for permanent
//	 spells you cast."
//
// The second line is ADR 0118 §3's granted offer with a spell filter
// and an energy price (ADR 0129 §5, PayEightEnergyForPermanentSpellsYouCast):
// an alternative cost for every permanent spell Nissa's controller casts,
// claimable wherever the printed mana cost could be paid (CR 118.9a),
// offered only to a caster with eight energy (CR 118.3) and paid through
// the one energy payer. A commander cast this way still pays its tax
// (CR 903.8). Energy is never waived (ADR 0129 §4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00037840-6089-42ec-8c5c-281f9f474504",
		Name:         "Nissa, Worldsoul Speaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Nissa, Worldsoul Speaker — you get {E}{E}", Do(GetEnergy{N: 2})),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayEightEnergyForPermanentSpellsYouCast()},
	})
}
