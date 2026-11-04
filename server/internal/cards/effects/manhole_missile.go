package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manhole Missile — {1}{R} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Manhole Missile deals 3 damage to target creature. You may put a card
//	 from your hand on the bottom of your library. If you do, draw a
//	 card."
//
// Two steps, in printed order: the damage, then an "up to one" pick from
// the hand (an empty hand is not asked). A picked card is put on the
// bottom of its owner's library, and the draw follows only if it got
// there: a commander its owner sends to the command zone instead
// (CR 903.9b) was not put on the bottom, so there is no draw (CR 608.2c).
// A target gone by resolution makes the spell do nothing at all
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c596b847-4e05-4279-8675-b4e39cb3dc91",
		Name:         "Manhole Missile",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DealDamageToTheTarget(3)(ctx.Game, ctx.Item)
			}),
			Instruction(manholeMissileMayTuck),
		),
	})
}

// manholeMissileMayTuck is "You may put a card from your hand on the
// bottom of your library": an up-to-one pick from the caster's hand.
func manholeMissileMayTuck(ctx *Context) error {
	player := ctx.Controller()
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Manhole Missile — you may put a card from your hand on the bottom of your library. If you do, draw a card",
		Cards:    hand,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  manholeMissileTucked,
	})
	return nil
}

var (
	// manholeMissileTucked is the pick's answer, keyed (#4253): the picked
	// card goes to the bottom of its owner's library, and the draw is told
	// whether it got there. An on-disk key.
	manholeMissileTucked = game.RegisterChooseCardsThen("choose-cards/manhole-missile-bottom", func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{r.Picked[0]}, game.TuckOptions{ToBottom: true},
			game.KeyedTail{Ref: manholeMissileDrawIfTucked, Player: r.Chooser, Source: r.Source})
	})

	// manholeMissileDrawIfTucked is "If you do, draw a card": Tail.Player
	// draws only when the card reached a library. An on-disk key.
	manholeMissileDrawIfTucked = game.RegisterTail("then/manhole-missile-draw-if-tucked", func(g *game.Game, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return g.DrawNForEffect(o.Tail.Player, 1)
	})
)
