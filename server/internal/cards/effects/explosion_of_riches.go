package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Explosion of Riches — Sorcery {5}{R} (#4052):
//
//	"Draw a card, then each other player may draw a card. Whenever a
//	 card is drawn this way, Explosion of Riches deals 5 damage to
//	 target opponent chosen at random from among your opponents."
//
// You draw first. Then each other player, in turn order from you, is
// asked one yes/no question: may they draw. The question is a KEYED
// option pick (OptionPickThen, Carry the players still to ask), so a
// table waiting on an opponent's answer is still a restore point.
//
// "Whenever a card is drawn this way" is one trigger per card actually
// drawn - yours and each accepted draw. A draw a replacement turned
// into something else, and a draw from an empty library, put no card
// into a hand and so trigger nothing; the card counts the cards that
// reached the player's hand (len of Game.DrawnThisTurn) rather than the
// answers given. Each trigger is a CR 603.12-style follow-up that goes
// on the stack after the spell resolves, with its OWN target: an
// opponent of the spell's controller drawn uniformly at random as it
// goes on the stack (ChosenAtRandom, #1815, CR 603.3d), so two triggers
// can name two different opponents. The damage is dealt by the spell,
// last-known from the graveyard.
//
// A player a draw limit stops (CR 121.2b) is not asked: they can't
// choose to draw (CanChooseToDrawForEffect). A player who has left the
// game is skipped. An answer that never comes (the chooser left) is a
// no.
//
// No simplification.
var (
	explosionOfRichesAnswer game.OptionPickThen

	explosionOfRichesDamageBody = game.ReflexiveBody("explosion-of-riches/damage",
		simpleBody(explosionOfRichesDamage),
		constTargets(func() *game.TargetSpec {
			return ChosenAtRandom(TargetPlayer("target opponent chosen at random from among your opponents", Opponent()))
		}))
)

func init() {
	// Registered here, not in the var block: the continuation asks the
	// next player, which names the key again (an initialization cycle
	// otherwise, as in Winter's Chill).
	explosionOfRichesAnswer = OptionPickThen("option-pick/explosion-of-riches-draw", explosionOfRichesAnswered)
	Register(Spec{
		OracleID:     "ffd469b4-413d-49f5-a873-e3f654240b31",
		Name:         "Explosion of Riches",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		// The opponents are asked after a prompt your draw asks (a dredge offer; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return explosionOfRichesDraw(ctx, ctx.Controller())
			}),
			Instruction(func(ctx *Context) error {
				return explosionOfRichesAsk(ctx, opponentsInTurnOrder(ctx.Game, ctx.Controller()))
			}),
		),
	})
}

// explosionOfRichesDraw has `player` draw a card and, if one reached
// their hand, puts one damage trigger on its way.
func explosionOfRichesDraw(ctx *Context, player uuid.UUID) error {
	before := len(ctx.Game.DrawnThisTurn[player])
	if err := (DrawCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if len(ctx.Game.DrawnThisTurn[player]) <= before {
		return nil
	}
	return WhenYouDo("Explosion of Riches - 5 damage to an opponent chosen at random", explosionOfRichesDamageBody).Apply(ctx)
}

// explosionOfRichesAsk asks the first of `order` who can choose to
// draw, carrying the rest.
func explosionOfRichesAsk(ctx *Context, order []uuid.UUID) error {
	for i, who := range order {
		if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
			continue
		}
		if !ctx.Game.CanChooseToDrawForEffect(who, 1) {
			continue
		}
		return PickOption{
			Player:   who,
			Question: "Explosion of Riches - draw a card?",
			Options: []game.ChoiceOption{
				{Label: "Don't draw a card"},
				{Label: "Draw a card"},
			},
			ThenKey: explosionOfRichesAnswer,
			Carry:   append([]uuid.UUID(nil), order[i+1:]...),
		}.Apply(ctx)
	}
	return nil
}

// explosionOfRichesAnswered is the keyed continuation: draw on a yes,
// then ask whoever is left.
func explosionOfRichesAnswered(ctx *Context, r game.OptionPicked) error {
	if r.Option != nil && r.Index == 1 {
		if err := explosionOfRichesDraw(ctx, r.Chooser); err != nil {
			return err
		}
	}
	return explosionOfRichesAsk(ctx, r.Carry)
}

// explosionOfRichesDamage is the trigger: the spell deals 5 damage to
// the opponent the engine drew.
func explosionOfRichesDamage(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 5}.Apply(NewContext(g, item))
}
