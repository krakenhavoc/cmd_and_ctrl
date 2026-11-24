package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Dreadlord — {3}{B}{B} Creature — Zombie Elf 3/3, Commander Legends (#5507):
//
//	"Deathtouch
//	 When this creature dies, non-Elf creatures get -3/-3 until end of turn.
//	 Encore {5}{B}{B}"
//
// The dies trigger locks in the non-Elf creatures on the battlefield as it
// resolves (CR 611.2c), friend or foe, and spares every Elf. Encore is the
// shared constructor (encore.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51b9d8ea-b0df-4a57-804e-6d395e301242",
		Name:            "Elvish Dreadlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []effects.ActivatedAbility{effects.Encore("{5}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Elvish Dreadlord — non-Elf creatures get -3/-3 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.Not(effects.Subtype("Elf"))), Power: -3, Toughness: -3,
					Label: "Elvish Dreadlord — -3/-3 until end of turn"})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3, Partial: true}}),
		},
	})
}
