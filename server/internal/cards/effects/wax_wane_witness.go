package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wax-Wane Witness — Creature — Bat Cleric {3}{W}, 2/4 (Bloomburrow,
// #4467):
//
//	"Flying, vigilance
//	 Whenever you gain or lose life during your turn, this creature gets
//	 +1/+0 until end of turn."
//
// Moonstone Harbinger's trigger without its once-each-turn limit: a gain
// is a positive EventChangeLife, a loss a negative one or damage that
// cost you life (s22PlayerLostLife). It fires only while you are the
// active player, once for each gain or loss. The +1/+0 goes on this
// object only, so a Witness that has left gets nothing.
//
// No simplification.
func init() {
	const label = "Wax-Wane Witness — this creature gets +1/+0 until end of turn"
	Register(Spec{
		OracleID:        "426dbe89-0cbb-4a3b-8812-d9761d063373",
		Name:            "Wax-Wane Witness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			Key:       label,
			AppliesTo: waxWaneWitnessLifeChangedOnYourTurn,
			Effect:    thisCreatureUntilEOT("Wax-Wane Witness — +1/+0", 1, 0),
		}},
	})
}

// waxWaneWitnessLifeChangedOnYourTurn is the trigger condition: it is
// your turn, and you gained or lost life.
func waxWaneWitnessLifeChangedOnYourTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	you := source.Controller
	if !IsYourTurn(g, you) {
		return false
	}
	if _, lost := s22PlayerLostLife(ev, you, g); lost {
		return true
	}
	return YouGainedLife(ev, source, lki, g)
}
