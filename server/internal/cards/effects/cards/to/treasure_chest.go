package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasure Chest — Artifact {3} (Adventures in the Forgotten Realms, S82):
//
//	"{4}, Sacrifice this artifact: Roll a d20.
//	 1 | Trapped! — You lose 3 life.
//	 2—9 | Create five Treasure tokens.
//	 10—19 | You gain 3 life and draw three cards.
//	 20 | Search your library for a card. If it's an artifact card, you may
//	 put it onto the battlefield. Otherwise, put that card into your hand.
//	 Then shuffle."
//
// The sacrifice is part of the cost, so the Chest is gone before the d20 and
// nothing here reads it. Row 20 is a search for any card (never optional: an
// unrestricted search takes a card when there is one) that leaves the card on
// top of the library, unshuffled, and the tail then decides: an artifact is a
// "you may" put onto the battlefield (a decline leaves it in the library), any
// other card goes to the hand, and the shuffle comes last either way, so the
// library ends in the printed state.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3df26058-c58d-4869-ae32-922d28c3aca2",
		Name:         "Treasure Chest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, Sacrifice this artifact: Roll a d20.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.RollDice{Sides: 20, N: 1, Then: treasureChestRolled}),
		}},
	})
}

// treasureChestRolled is the results table, after the d20. An on-disk key.
var treasureChestRolled = effects.DiceTail("dice/treasure-chest-rolled", func(ctx *effects.Context, rolls []int) error {
	g, me := ctx.Game, ctx.Controller()
	switch r := rolls[0]; {
	case r == 1:
		return g.ChangePlayerLifeForEffect(ctx.Source(), me, -3)
	case r <= 9:
		return effects.TreasureTokens(g, me, 5)
	case r <= 19:
		if err := (effects.GainLife{Player: me, Amount: 3}).Apply(ctx); err != nil {
			return err
		}
		return effects.DrawCards{Player: me, N: 3}.Apply(ctx)
	}
	return effects.SearchLibrary{
		Player: me,
		Dest:   game.ZoneLibrary,
		ToTop:  true,
		Limit:  1,
		Reason: "Treasure Chest — search your library for a card",
		// The shuffle is the tail's, after the card has gone where it goes.
		ThenTail: ctx.Tail(treasureChestFound),
	}.Apply(ctx)
})

// treasureChestFound is the rest of the search, told the card found (on top
// of the library): an artifact may go onto the battlefield, anything else to
// the hand, then shuffle. An on-disk key.
var treasureChestFound = effects.CardTail("dice/treasure-chest-found", func(ctx *effects.Context, o game.TailOutcome) error {
	g, me := ctx.Game, ctx.Controller()
	if len(o.IDs) == 0 {
		return g.ShuffleLibraryForEffect(me)
	}
	c, ok := g.LookupCardForEffect(o.IDs[0])
	if !ok {
		return g.ShuffleLibraryForEffect(me)
	}
	if c.IsArtifact() {
		return effects.PutFromLibraryOntoBattlefield{
			Player:   me,
			Cards:    o.IDs,
			Max:      1,
			Optional: true,
			Label:    "Treasure Chest — you may put the artifact card onto the battlefield",
			ThenTail: effects.ThenShuffle,
		}.Apply(ctx)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(me, o.IDs, ctx.Tail(effects.ThenShuffle))
})
