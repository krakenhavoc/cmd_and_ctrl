package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Glory's Rise — Creature — Angel {5}{W}{W}, 4/6:
//
//	"Flying
//	 When this creature enters, exile all Zombies, then return all Human
//	 creature cards from your graveyard to the battlefield."
//
// Every Zombie permanent, whoever controls it, is exiled (not destroyed:
// no dies triggers, indestructible does not save it). Then every Human
// creature card in its controller's graveyard returns in one simultaneous
// move (CR 608.2f), after the exile has finished. "Zombies" and "Human"
// are words of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6a606af-ea08-4cd9-bfd9-60d76e232e85",
		Name:            "Angel of Glory's Rise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Angel of Glory's Rise — exile all Zombies, then return all Human creature cards from your graveyard to the battlefield",
				effects.Do(
					effects.ExileAllMatching{Match: effects.Subtype("Zombie")},
					effects.Instruction(func(ctx *effects.Context) error {
						human := ctx.TextSubtype("Human")
						return effects.ReturnFromGraveyardTogether{
							Targets: effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
								return c.IsCreature() && c.HasSubtype(human)
							}),
							Controller: ctx.Controller(),
						}.Apply(ctx)
					}),
				)),
		},
	})
}
