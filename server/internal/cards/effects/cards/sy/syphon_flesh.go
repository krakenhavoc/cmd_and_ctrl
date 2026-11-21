package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syphon Flesh — Sorcery {4}{B}:
//
//	"Each other player sacrifices a creature of their choice. You
//	 create a 2/2 black Zombie creature token for each creature
//	 sacrificed this way."
//
// An edict with a tail (EachPlayerSacrifices.Then): each other player
// picks their own creature, and only once every seat has answered and
// the creatures have gone are the Zombies counted — one per creature
// that really left. A player with no creature is not asked and adds
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc7c3151-ad07-429a-8af8-e4f24b7d9f6b",
		Name:         "Syphon Flesh",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Creature(),
				Label:            "a creature",
				Keyed:            SyphonFleshEdicted,
			}.Apply(ctx)
		},
	})
}

// syphonFleshEdicted is a Zombie for each creature sacrificed this way,
// keyed (#4290).
var SyphonFleshEdicted = effects.RegisterEdictThen("prompt-run/syphon-flesh-zombies", func(ctx *effects.Context, sacrificed game.PromptedSacrifices) error {
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: sacrificed.Count()}.Apply(ctx)
})
