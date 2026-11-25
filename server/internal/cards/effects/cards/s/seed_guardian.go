package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seed Guardian — Creature — Elemental {2}{G}{G}, 3/4 (Oath of the
// Gatewatch, S82 OLD #5545):
//
//	"Reach
//	 When this creature dies, create an X/X green Elemental creature token,
//	 where X is the number of creature cards in your graveyard."
//
// X is counted as the trigger resolves (CR 608.2h reads the board as it is),
// so the Guardian, already in the graveyard, counts itself. It is the size
// the token enters with: the printed Elemental token is */*, so the template
// is the 0/0 row every X-sized Elemental shares (Marath, Will of the Wild).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "728045bf-0959-44ff-8de9-16ec35fd248f",
		Name:            "Seed Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Seed Guardian — create an X/X green Elemental creature token, where X is the number of creature cards in your graveyard",
				effects.Do(effects.Instruction(seedGuardianElemental))),
		},
	})
}

func seedGuardianElemental(ctx *effects.Context) error {
	x := effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
	tok := ctx.TextToken(effects.TokenCard("0/0 green Elemental"))
	tok.Power, tok.Toughness = x, x
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
}
