package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brawl-Bash Ogre — Creature — Ogre Warrior {2}{B}{R}, 3/3 (Core Set 2019):
//
//	"Menace
//	 Whenever this creature attacks, you may sacrifice another creature. If
//	 you do, this creature gets +2/+2 until end of turn."
//
// Kor, Cause of Vengeance's optional sacrifice: asked as the trigger resolves
// over the other creatures its controller controls then (CR 608.2c); declining
// or having none does nothing, and "if you do" follows only a creature that
// really left. The pump lands only if the Ogre is still the permanent that
// attacked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0156bfbc-26bc-4699-a7ff-437f7c043716",
		Name:            "Brawl-Bash Ogre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Brawl-Bash Ogre — you may sacrifice another creature; if you do, it gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Brawl-Bash Ogre — sacrifice another creature to get +2/+2 until end of turn?",
						effects.Creature(), true, brawlBashOgreSacrificed)
				}),
		},
	})
}

// brawlBashOgreSacrificed is "if you do": the Ogre gets +2/+2 until end of
// turn. An on-disk key: never renamed, never reused.
var brawlBashOgreSacrificed = effects.RegisterSacrificedOne("prompt-run/brawl-bash-ogre-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: 2, Toughness: 2,
		Label: "Brawl-Bash Ogre — +2/+2 until end of turn"}.Apply(ctx)
})
