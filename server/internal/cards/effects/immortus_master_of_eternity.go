package effects

import (
	"github.com/google/uuid"

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
	powerUp := PowerUp("{5}{U}{U}: Each player shuffles their hand and graveyard into their library, then draws seven cards. Put a +1/+1 counter on Immortus.",
		ManaCost("{5}{U}{U}"), func(g *game.Game, item *game.StackItem) error {
			players := tablePlayers(NewContext(g, item))
			if err := echoOfEonsShuffleEach(g, players, players); err != nil {
				return err
			}
			return plusOneCountersOnThis(1)(g, item)
		})
	powerUp.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:     "8952a7e6-4ede-42bb-b90c-00ea4f24f8f8",
		Name:         "Immortus, Master of Eternity",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: immortusProduced,
			Restrictions:   []string{ManaRestrictCast, ManaRestrictNotType("Creature")},
			Label:          "Add {U} for each card you've drawn this turn. Spend this mana only to cast noncreature spells",
		}},
		Activated: []ActivatedAbility{powerUp},
	})
}

// immortusProduced is one {U} per card its controller has drawn this turn.
func immortusProduced(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
	return game.ManaAmount{Options: []string{"U"}, N: g.TurnTallyFor(controller).CardsDrawn}
}
