package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gev, Scaled Scorch — Legendary Creature — Lizard Mercenary {B}{R}, 3/2
// (Bloomburrow, #4467):
//
//	"Ward—Pay 2 life.
//	 Other creatures you control enter with an additional +1/+1 counter
//	 on them for each opponent who lost life this turn.
//	 Whenever you cast a Lizard spell, Gev deals 1 damage to target
//	 opponent."
//
// The ward is Hexing Squelcher's (Ward with WardLife). The counters are
// Renata's CR 614.12 entry replacement over the other creatures
// entering under Gev's controller, counted as each one enters: one per
// opponent still in the game whose turn tally shows life lost this turn
// (a player who lost life twice counts once). Gev itself is never in the
// set it modifies. The cast trigger is Harnesser of Storms' "whenever you
// cast" over a spell with the Lizard subtype, read through Gev's text (CR
// 612.2), with Abraded Bluffs' "target opponent".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77b832a8-a3af-44fe-bc67-72510d292667",
		Name:         "Gev, Scaled Scorch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Gev, Scaled Scorch — ward, pay 2 life"),
			Targeting(WheneverYouCast(Subtype("Lizard"), "Gev, Scaled Scorch — 1 damage to target opponent",
				DamageToFirstTarget(1)),
				TargetPlayer("target opponent", Opponent())),
		},
		Replacements: []game.ReplacementEffect{gevOthersEnterWithCounters()},
	})
}

// gevOthersEnterWithCounters is Renata's entry replacement over the other
// creatures entering under Gev's controller
// (b27OtherCreaturesYouControlEnterWithACounter), with the count read as
// each one enters: one +1/+1 counter per opponent who lost life this
// turn, and none when no opponent has.
func gevOthersEnterWithCounters() game.ReplacementEffect {
	r := b27OtherCreaturesYouControlEnterWithACounter("Gev, Scaled Scorch: other creatures you control enter with an additional +1/+1 counter for each opponent who lost life this turn")
	r.Replace = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
		if n := gevOpponentsWhoLostLifeThisTurn(g, src.Controller); n > 0 {
			ev.AddCounterAtETB(game.CounterPlusOne, n)
		}
		return nil
	}
	return r
}

// gevOpponentsWhoLostLifeThisTurn counts the opponents of `player` still
// in the game who lost any life this turn.
func gevOpponentsWhoLostLifeThisTurn(g *game.Game, player uuid.UUID) int {
	n := 0
	eachOpponent(g, player, func(opponent uuid.UUID) bool {
		if g.TurnTallyFor(opponent).LifeLost > 0 {
			n++
		}
		return false
	})
	return n
}
