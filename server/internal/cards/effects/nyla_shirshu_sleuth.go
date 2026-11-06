package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyla, Shirshu Sleuth — Legendary Creature — Mole Beast {4}{B}, 4/5
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"When Nyla enters, exile up to one target creature card from your
//	 graveyard. If you do, you lose X life and create X Clue tokens,
//	 where X is that card's mana value.
//	 At the beginning of your end step, if you control no Clues, return
//	 target card exiled with Nyla to its owner's hand."
//
// The exile stamps the CR 607.2a link to Nyla as the object she is now
// (Quintorius, Loremaster's shape), and "if you do" is the exile's tail,
// told the card only when it reached exile; X is its mana value there.
// Making Clues this way is not investigating (the ruling).
//
// The end-step "if" is intervening (CR 603.4): with a Clue on the
// battlefield under your control the trigger does not go on the stack, and
// one made in response leaves it doing nothing. Its target is a card
// exiled with this Nyla only (ExiledWithThis), so a Nyla that left and
// came back has no claim on her earlier self's exiles (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3492cee8-b854-449a-ada4-a4e8a4333d87",
		Name:         "Nyla, Shirshu Sleuth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Nyla, Shirshu Sleuth — exile up to one target creature card from your graveyard; you lose X life and create X Clues",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ExileTarget{Target: id, Linked: true, ThenTail: ctx.Tail(nylaShirshuSleuthExiled)}.Apply(ctx)
				}),
				TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1)),
			Targeting(On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && !nylaControlsAClue(g, source.Controller)
			}, "Nyla, Shirshu Sleuth — you control no Clues: return target card exiled with Nyla to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if nylaControlsAClue(g, item.Controller) {
						return nil
					}
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BounceToHand{Target: id}.Apply(ctx)
				}),
				ExiledWithThis(TargetCardInExile("target card exiled with Nyla, Shirshu Sleuth"))),
		},
	})
}

// nylaShirshuSleuthExiled is "if you do, you lose X life and create X Clue
// tokens", told the exiled card. An on-disk key: never renamed, never
// reused.
var nylaShirshuSleuthExiled = CardTail("nyla-shirshu-sleuth/lose-x-make-x-clues", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	x := c.ManaValue()
	if x <= 0 {
		return nil
	}
	if err := (GainLife{Amount: -x}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Template: ClueToken(), N: x}.Apply(ctx)
})

// nylaControlsAClue reports whether the player controls a Clue, read
// through the layered view.
func nylaControlsAClue(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype("Clue") {
			return true
		}
	}
	return false
}
