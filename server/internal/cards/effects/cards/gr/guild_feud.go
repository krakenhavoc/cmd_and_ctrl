package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guild Feud — Enchantment {5}{R} (Return to Ravnica, #5566):
//
//	"At the beginning of your upkeep, target opponent reveals the top three
//	 cards of their library, may put a creature card from among them onto the
//	 battlefield, then puts the rest into their graveyard. You do the same
//	 with the top three cards of your library. If two creatures are put onto
//	 the battlefield this way, those creatures fight each other."
//
// Two put-from-library prompts in a row (put_from_library.go): the opponent's,
// whose answer is theirs to give, then yours. The first tail puts the
// opponent's rest into their graveyard and opens yours, carrying the creature
// the opponent put (a nil ID when they put none, so the revealed cards keep
// their place in the carry); the second puts your rest into your graveyard and,
// when both of you put a creature, has them fight (CR 701.14). Both tails ask
// for the trigger's own item so the controller stays you throughout.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc5d0e18-2d10-47df-9742-2ea6c6d44431",
		Name:         "Guild Feud",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Guild Feud — target opponent reveals the top three cards, may put a creature card onto the battlefield and puts the rest into their graveyard; you do the same; if two creatures were put onto the battlefield, they fight",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
							return nil
						}
						return effects.PutFromLibraryOntoBattlefield{
							Player:   t.ID,
							Cards:    g.RevealTopOfLibraryForEffect(t.ID, ctx.Source(), 3, "Guild Feud — target opponent reveals the top three cards of their library"),
							Match:    effects.Creature(),
							Max:      1,
							Optional: true,
							Label:    "Guild Feud — you may put a creature card from among them onto the battlefield",
							ThenTail: guildFeudOpponentTail,
							ThenItem: true,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// guildFeudOpponentTail is the end of the opponent's half and the start of
// yours. An on-disk key: never renamed, never reused.
var guildFeudOpponentTail = effects.PutFromLibraryTail("put-from-library/guild-feud-opponent", func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
	if err := effects.PutRestIntoGraveyard(ctx.Game, res); err != nil {
		return err
	}
	me := ctx.Controller()
	theirs := uuid.Nil
	if len(res.Entered) > 0 {
		theirs = res.Entered[0]
	}
	return effects.PutFromLibraryOntoBattlefield{
		Player:    me,
		Cards:     ctx.Game.RevealTopOfLibraryForEffect(me, ctx.Source(), 3, "Guild Feud — reveal the top three cards of your library"),
		Match:     effects.Creature(),
		Max:       1,
		Optional:  true,
		Label:     "Guild Feud — you may put a creature card from among them onto the battlefield",
		ThenTail:  guildFeudYourTail,
		ThenCarry: []uuid.UUID{theirs},
		ThenItem:  true,
	}.Apply(ctx)
})

// guildFeudYourTail puts your rest into your graveyard, then makes the two
// creatures fight if both halves put one. An on-disk key: never renamed,
// never reused.
var guildFeudYourTail = effects.PutFromLibraryTailCarrying("put-from-library/guild-feud-you", 1, func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
	if err := effects.PutRestIntoGraveyard(ctx.Game, res); err != nil {
		return err
	}
	if len(res.Carry) == 0 || res.Carry[0] == uuid.Nil || len(res.Entered) == 0 {
		return nil
	}
	return effects.B10Fight(ctx, res.Carry[0], res.Entered[0])
})
