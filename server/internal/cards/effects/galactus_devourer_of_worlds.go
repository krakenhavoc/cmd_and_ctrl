package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galactus, Devourer of Worlds — Legendary Creature — Elder Alien {10},
// 12/12:
//
//	"Flying, trample, indestructible
//	 When Galactus enters, exile target permanent.
//	 Insatiable Hunger — Galactus attacks an opponent with the most life
//	 among your opponents each combat if able unless you control a
//	 creature named Silver Surfer, Galactus's Herald."
//
// The enters trigger is Duplicant's exile without the imprint. Insatiable
// Hunger is an ability word, so the line is an ordinary static: a CR
// 508.1d requirement obeyed only by attacking one of the opponents tied
// for the most life (#2744), switched off while its controller controls
// a creature named Silver Surfer, Galactus's Herald.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4232995b-68c0-4514-8cde-bc62b9d1cbaa",
		Name:            "Galactus, Devourer of Worlds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "indestructible"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPermanent("target permanent"),
			Key:       "Galactus, Devourer of Worlds — exile target permanent",
			Effect:    b27ExileChosenTarget,
		}},
		Static: []game.StaticAbility{
			AttacksAnOpponentWithTheMostLifeEachCombat(youControlACreatureNamed("Silver Surfer, Galactus's Herald")),
		},
	})
}
