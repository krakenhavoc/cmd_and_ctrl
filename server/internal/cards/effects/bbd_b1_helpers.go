package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bbd_b1_helpers.go — the clauses the Battlebond batch (#5529) shares.
//
// "Your team" is "you" (CR 102.4): in a game that isn't a multiplayer
// game between teams, "your team" means the same thing as "you", and this
// engine has no team format, so "your team controls" is "you control" at
// every table it runs.

// bbdWheneverAttacksWithAnotherWarrior is "Whenever this creature attacks,
// if your team controls another Warrior, <effect>" (Aurora Champion,
// Bull-Rush Bruiser, Sickle Dancer). The "if" is an intervening one (CR
// 603.4): another Warrior is checked as the creature attacks and again as
// the trigger resolves, through the layers, so a changeling counts and
// losing the other Warrior in response stops the effect. "Warrior" is a
// word of the source's text (CR 612.2).
func BbdWheneverAttacksWithAnotherWarrior(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondBoard{
		Name: "your team controls another Warrior",
		Test: func(ctx *Context) bool {
			return MshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Warrior"))
		},
	}, On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ThisAttacked(ev, source, lki, g) && MshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Warrior"))
	}, label, effect))
}

// bbdCounteredCreatureAttacksAnOpponent is Soulblade Corrupter's "a
// creature with a +1/+1 counter on it attacks one of your opponents": any
// player's creature, judged as it is declared, attacking a player (not a
// planeswalker or battle) other than the source's controller.
func BbdCounteredCreatureAttacksAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	opp := AttackedPlayer(g, ev)
	if opp == uuid.Nil || opp == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
}

// bbdAnotherNontokenCreatureEntered is Bramble Sovereign's "another
// nontoken creature enters": anyone's, not only the source controller's.
func BbdAnotherNontokenCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !IsToken(c)
}

// bbdAPlayerWonACoinFlip is "whenever a player wins a coin flip" (Okaun,
// Eye of Chaos; Zndrsplt, Eye of Wisdom): YouWonACoinFlip for any flipper.
// Only the player who flips a coin wins it (CR 705.2), and a face-only flip
// carries Won false and never matches.
func BbdAPlayerWonACoinFlip(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventFlipCoin && ev.Won && ev.Actor != uuid.Nil
}

// bbdFlipUntilYouLose is "flip a coin until you lose a flip" (Okaun and
// Zndrsplt's beginning of combat): one called flip at a time, each won flip
// asking the next, and the first lost flip ending the chain. What winning
// does is whatever triggers on the EventFlipCoin each flip emits. The
// question rides the call as its one word so the chain keeps the card's
// name.
func BbdFlipUntilYouLose(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Controller == uuid.Nil {
			return nil
		}
		bbdFlipAgain(g, item.Controller, item.SourceCardID, question, 0)
		return nil
	}
}

// bbdFlipAgain queues one link of the chain.
func bbdFlipAgain(g *game.Game, flipper, source uuid.UUID, question string, wins int) {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  flipper,
		Source:   source,
		Question: question,
		Wins:     wins,
		ThenKey:  bbdFlipChained,
		Words:    []string{question},
	})
}

// bbdFlipChained is the settled call: a won flip asks the next. An on-disk
// key, registered in init because its body queues itself.
var bbdFlipChained game.CoinFlipThen

func init() {
	bbdFlipChained = CoinFlipThen("coin-flip/bbd-flip-until-you-lose", func(ctx *Context, r game.CoinFlipped) error {
		if !WonFirstFlip(r) {
			return nil
		}
		bbdFlipAgain(ctx.Game, r.Flipper, r.Source, CarriedWord(r.Words, 0), r.Wins+1)
		return nil
	})
}

// bbdCastBanOfTheChosenType is Archon of Valor's Reach's "Players can't
// cast spells of the chosen type": every player, the Archon's controller
// included, is refused a spell with the card type the Archon's answer
// names, read off the spell as it is announced. Unanswered, nothing is
// refused.
func BbdCastBanOfTheChosenType(label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			chosen := q.Source.ChosenOption
			return chosen != "" && q.Card.HasCardType(strings.ToLower(chosen))
		},
	}
}
