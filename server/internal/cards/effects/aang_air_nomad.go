package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, Air Nomad — Legendary Creature — Human Avatar Ally {3}{W}{W}, 5/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying
//	 Vigilance
//	 Other creatures you control have vigilance."
//
// Its own keywords ride PrintedKeywords; the grant is a layer-6 keyword
// grant to every other creature you control (CR 613.1f), read through
// the layers so a creature that becomes yours later has it too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e46b2ac6-97d6-48f2-aba3-cd92f8e092c6",
		Name:            "Aang, Air Nomad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{
			KeywordGrant(OtherCreaturesYouControl, "vigilance"),
		},
	})
}
