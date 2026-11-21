package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Redcap Gutter-Dweller — Creature — Goblin Warrior {2}{R}{R}, 3/3 (Wilds
// of Eldraine, #4471):
//
//	"Menace
//	 When this creature enters, create two 1/1 black Rat creature tokens
//	 with "This token can't block."
//	 At the beginning of your upkeep, you may sacrifice another creature.
//	 If you do, put a +1/+1 counter on this creature and exile the top
//	 card of your library. You may play that card this turn."
//
// The entry is Harried Spearguard's Rat twice. The upkeep asks as it
// resolves (Bolg's maySacrificeOne) over every other creature you control;
// "if you do" runs only for a creature that really left. The counter goes on
// the Dweller only while it is still the permanent that triggered, but the
// exile and the permission are the card's own sentence and happen even when
// it has left (the ruling). The exiled card may be played this turn
// (Breeches' impulse): a land only in a main phase with the stack empty and
// a land drop, a spell with its costs and timing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ada48353-4c86-44d6-83a5-e7e3c703b948",
		Name:            "Redcap Gutter-Dweller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Redcap Gutter-Dweller — create two 1/1 black Rat creature tokens with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 2})),
			effects.AtYourUpkeep("Redcap Gutter-Dweller — you may sacrifice another creature; if you do, put a +1/+1 counter on it and exile the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Redcap Gutter-Dweller — sacrifice another creature to grow and exile the top card of your library?",
						effects.Creature(), true, redcapSacrificed)
				}),
		},
	})
}

// redcapSacrificed is the "if you do", keyed (#4290).
var redcapSacrificed = effects.RegisterSacrificedOne("prompt-run/redcap-gutter-dweller-sacrificed", func(ctx *effects.Context, _ uuid.UUID) error {
	if effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ExileTopCardYouMayPlayThisTurn(ctx.Item, ctx, 0)
})
