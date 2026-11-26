package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossus of Akros — Artifact Creature — Golem {8}, 10/10 (Theros):
//
//	"Defender, indestructible
//	 {10}: Monstrosity 10. (If this creature isn't monstrous, put ten +1/+1
//	 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has trample and can attack as
//	 though it didn't have defender."
//
// Monstrosity is the shared activated ability. While monstrous the Colossus has
// trample (MonstrousKeywords) and the defender waiver is a self-only static
// that holds behind the same Monstrous gate (CR 508.1c: it may attack, the
// keyword stays).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96d508a6-ecd2-4a97-8802-df099a372084",
		Name:            "Colossus of Akros",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "indestructible"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{10}"), 10)},
		Static: []game.StaticAbility{
			effects.MonstrousKeywords("trample"),
			effects.OldB396MonstrousMayAttackDespiteDefender(),
		},
	})
}
