package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ogre Chitterlord — Creature — Ogre Warrior {4}{R}{R}, 6/5 (Wilds of
// Eldraine, #4471):
//
//	"Menace
//	 Whenever this creature enters or attacks, create two 1/1 black Rat
//	 creature tokens with "This token can't block." Then if you control
//	 five or more Rats, each Rat you control gets +2/+0 until end of
//	 turn."
//
// One ability, two conditions (Sun Titan's WhenThisEntersOrAttacks).
// The Rats are Twisted Sewer-Witch's token. The count is taken after
// both tokens exist, and so is the set of Rats that get +2/+0: it is
// read once (CR 611.2c), so a Rat that enters later in the turn gets
// nothing. A changeling is a Rat (CR 702.73a), and "Rat" is a word of
// the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "532cac99-e3cf-4869-a19f-cba67a2a6fd0",
		Name:            "Ogre Chitterlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Ogre Chitterlord — create two Rat tokens, then each Rat you control gets +2/+0 if you control five or more",
				Do(
					// The pump waits for a prompt the tokens' creation asks (CR 608.2c, #4155).
					CreateToken{Template: CantBlockRatToken(), N: 2},
					Instruction(func(ctx *Context) error {
						rat := ctx.TextSubtype("Rat")
						yourRat := func(_ *game.Game, you uuid.UUID, c game.Card) bool {
							return c.Controller == you && c.HasSubtype(rat)
						}
						n := 0
						for _, c := range ctx.Game.BattlefieldCardsForEffect() {
							if yourRat(ctx.Game, ctx.Controller(), c) {
								n++
							}
						}
						if n < 5 {
							return nil
						}
						return BoostUntilEOT{Match: And(Creature(), yourRat), Power: 2, Label: "Ogre Chitterlord — +2/+0"}.Apply(ctx)
					}),
				)),
		},
	})
}
