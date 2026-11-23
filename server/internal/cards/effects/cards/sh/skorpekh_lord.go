package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skorpekh Lord — Artifact Creature — Necron Noble {2}{B}, 3/2 (Warhammer
// 40,000 Commander, #5948):
//
//	"Menace
//	 Command Protocols — Other artifact creatures you control get +1/+0 and
//	 have menace.
//	 Unearth {2}{B}"
//
// Mist Dancer's lord shape over other artifact creatures you control: a layer
// 7c +1/+0 and a layer 6 menace grant, read through the layers.
//
// No simplification.
func init() {
	applies := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.IsCreature() &&
			target.HasCardType("artifact") && target.Controller == source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:        "c40e8d52-d4cc-4502-8b51-a5612dbe1ada",
		Name:            "Skorpekh Lord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			effects.B16Anthem(applies, 1, 0),
			effects.B16GrantKeywords(applies, "menace"),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{B}")},
	})
}
