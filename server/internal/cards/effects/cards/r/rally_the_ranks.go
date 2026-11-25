package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally the Ranks — Enchantment {1}{W} (Kaldheim):
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1."
//
// The as-enters choice and the chosen-type lord of Icon of Ancestry, with
// nothing else on the card.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eeb07c8a-21d1-43fa-a5f3-ee3fe328a671",
		Name:         "Rally the Ranks",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Rally the Ranks"),
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
	})
}
