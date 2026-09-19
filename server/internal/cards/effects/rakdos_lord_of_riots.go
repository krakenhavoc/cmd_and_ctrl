package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos, Lord of Riots — {B}{B}{R}{R} 6/6 Legendary Demon:
//
//	"You can't cast this spell unless an opponent lost life this turn.
//	 Flying, trample
//	 Creature spells you cast cost {1} less to cast for each 1 life
//	 your opponents have lost this turn."
//
// The card #760 called misattributed, and it was right: Rakdos'
// restriction is on casting RAKDOS — he is not on the battlefield
// when it applies — so it is the legendary-sorcery shape, a
// CastCondition on the card, and not a CastRestriction contributed by
// a permanent. That the one gate covers both is why the two issues
// are one ADR (ADR 0073 §7).
//
// Checked at announce and never again, like every cast condition:
// life gained back after Rakdos is on the stack does not counter him.
//
// The third line is an ordinary cost modifier and has been
// expressible since ADR 0048 — it is here because leaving it out
// would ship a Rakdos strictly worse than printed (#259), and because
// it reads the same turn tally the condition does. Note what it is
// NOT: a reduction of Rakdos' own cost. "Creature spells YOU cast"
// is a static from the battlefield, so it applies to the creatures
// Rakdos enables, not to Rakdos.
func init() {
	Register(Spec{
		OracleID:        "143a269a-b9ee-48ba-bd7b-4aa46eb36778",
		Name:            "Rakdos, Lord of Riots",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return opponentLifeLostThisTurn(g, controller) > 0
		},
		CastConditionLabel: "You can't cast this spell unless an opponent lost life this turn.",
		CostModifiers: []game.CostModifier{
			CostsLessEach(
				func(q game.CostQuery) int {
					return opponentLifeLostThisTurn(q.Game, q.Controller)
				},
				"Creature spells you cast cost {1} less to cast for each 1 life your opponents have lost this turn.",
				YourSpell(), CreatureSpell()),
		},
	})
}

// opponentLifeLostThisTurn totals the life `player`'s opponents have
// lost this turn, off the engine's own per-turn tally.
//
// One helper, two readers on the same card — the cast condition asks
// whether it is more than zero and the cost reduction asks how much —
// so the two can never disagree about who counts as an opponent.
//
// Eliminated seats are still counted while they are seated: the life
// was lost this turn, and CR 800.4 removing a player later does not
// unhappen it.
func opponentLifeLostThisTurn(g *game.Game, player uuid.UUID) int {
	total := 0
	for _, seat := range g.Seats {
		if seat.ID == player {
			continue
		}
		total += g.TurnTallyFor(seat.ID).LifeLost
	}
	return total
}
