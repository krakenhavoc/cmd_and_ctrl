package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lecturing Scornmage — Creature — Human Warlock {B}, 1/1 (Secrets of
// Strixhaven, #4456):
//
//	"Repartee — Whenever you cast an instant or sorcery spell that targets
//	 a creature, put a +1/+1 counter on this creature."
//
// Repartee is an ability word; the ability is one trigger on your instant
// and sorcery casts, read from the targets the spell was cast with
// (informedInkwrightCast, heroic's reading): a target that is a creature on
// the battlefield as the spell is cast. It resolves before the spell, so
// the counter lands even if the spell is countered or its target dies, and
// several creature targets still make one counter. The counter goes on
// only if the Scornmage is still the permanent that triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29c23fd7-0ac4-46af-8ecd-92b726575252",
		Name:         "Lecturing Scornmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Lecturing Scornmage — Repartee: put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
