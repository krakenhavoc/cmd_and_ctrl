package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lux Artillery — Artifact {4} (EDHREC rank 3706):
//
//	"Whenever you cast an artifact creature spell, it gains sunburst.
//	 (It enters with a +1/+1 counter on it for each color of mana
//	 spent to cast it.)
//	 At the beginning of your end step, if there are thirty or more
//	 counters among artifacts and creatures you control, this
//	 artifact deals 10 damage to each opponent."
//
// The counters deck's alternate win. The end-step half is Lathiel's
// shape: the count — every counter of every kind on the artifacts
// and creatures the controller controls, a permanent that is both
// counted once — is a CR 603.4 intervening-if, checked when the
// trigger would go on the stack and again on resolution, so the
// trigger never appears below thirty and a board shrunk in response
// stops the damage. The Artillery is the damage source, so a
// Fog-class shield stops it.
//
// The sunburst half (ADR 0109 §11, #1552) is a cast trigger that gives
// the SPELL sunburst on the stack (ThatSpellGains). The spell carries it
// onto the permanent (CR 400.7a), and the engine counts it as the
// permanent enters, like a printed one: an Etched Oracle cast under the
// Artillery has two instances and counts the colours twice
// (CR 702.44d). A spell countered before the trigger resolves gains
// nothing.
//
// The grant arrives AFTER the spell's costs are paid, so the payment is
// not spread across colours for it — correct, because the colours are
// already spent by then. Lux Artillery declares
// ReadsYourCasts, so a waived artifact creature cast under it asks what
// was paid outside the pool and the colours are exact (#3855).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "cbd76b22-d04e-48e7-bcab-c7f5466d67d8",
		Name:           "Lux Artillery",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Artifact(), effects.Creature()), "Lux Artillery — it gains sunburst",
				effects.Do(effects.ThatSpellGains{Keywords: []string{game.KeywordSunburst}})),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B35EndStepAndThirtyCounters(ev, source, g)
			}, "Lux Artillery — 10 damage to each opponent", effects.B35TenDamageToEachOpponentIfThirtyCounters),
		},
	})
}
