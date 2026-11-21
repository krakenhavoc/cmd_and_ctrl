package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telim'Tor — Legendary Creature — Human Knight {4}{R}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 Whenever Telim'Tor attacks, all attacking creatures with flanking get
//	 +1/+1 until end of turn."
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). The attack trigger is Kangee, Sky Warden's mass pump with
// flanking in place of flying: the set is locked as it resolves
// (CR 611.2c), so only creatures that are attacking and have flanking
// then get +1/+1, Telim'Tor among them, and flanking gained or lost later
// changes nothing. "Attacking creatures" is every attacker, not only
// yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b13a723d-80a9-4116-9857-baf3b373f87a",
		Name:            "Telim'Tor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered: []game.TriggeredAbility{
			effects.RetiredFlankingRow("Telim'Tor"),
			effects.WheneverThisAttacks("Telim'Tor — all attacking creatures with flanking get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.AttackingCreature(), effects.HasKeyword(game.KeywordFlanking)), Power: 1, Toughness: 1,
					Label: "Telim'Tor — attacking creatures with flanking get +1/+1"})),
		},
	})
}
