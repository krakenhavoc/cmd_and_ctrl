package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immortus, Master of Eternity — Legendary Creature — Human Hero Villain
// {2}{U}{U}, 2/2 (Marvel Super Heroes Commander):
//
//	"{T}: Add {U} for each card you've drawn this turn. Spend this mana only
//	 to cast noncreature spells.
//	 Power-up — {5}{U}{U}: Each player shuffles their hand and graveyard into
//	 their library, then draws seven cards. Put a +1/+1 counter on Immortus.
//	 (Activate each power-up ability only once. Reduce the cost by his mana
//	 cost if he entered this turn.)"
//
// The mana ability is Nardole's shape: the amount is read when it is
// activated, from the controller's cards-drawn tally for the turn, and the
// spend clause is "cast, and not a creature spell". The power-up is Echo of
// Eons' wheel (every player's hand and graveyard tucked together, shuffled,
// then seven cards each), followed by the counter.
//
// No simplification.
func init() {
	powerUp := effects.PowerUp("{5}{U}{U}: Each player shuffles their hand and graveyard into their library, then draws seven cards. Put a +1/+1 counter on Immortus.",
		effects.ManaCost("{5}{U}{U}"), func(g *game.Game, item *game.StackItem) error {
			players := effects.TablePlayers(effects.NewContext(g, item))
			if err := effects.EchoOfEonsShuffleEach(g, players, players); err != nil {
				return err
			}
			return effects.PlusOneCountersOnThis(1)(g, item)
		})
	powerUp.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:     "8952a7e6-4ede-42bb-b90c-00ea4f24f8f8",
		Name:         "Immortus, Master of Eternity",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: immortusProduced,
			Restrictions:   []string{effects.ManaRestrictCast, effects.ManaRestrictNotType("Creature")},
			Label:          "Add {U} for each card you've drawn this turn. Spend this mana only to cast noncreature spells",
		}},
		Activated: []effects.ActivatedAbility{powerUp},
	})
}

// immortusProduced is one {U} per card its controller has drawn this turn.
func immortusProduced(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
	return game.ManaAmount{Options: []string{"U"}, N: g.TurnTallyFor(controller).CardsDrawn}
}
