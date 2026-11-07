package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zuko, Firebending Master — Legendary Creature — Human Noble Ally
// {1}{R}, 2/2 (Avatar: The Last Airbender Eternal, #5075):
//
//	"First strike
//	 Firebending X, where X is the number of experience counters you
//	 have. (Whenever this creature attacks, add X {R}. This mana lasts
//	 until end of combat.)
//	 Whenever you cast a spell during combat, you get an experience
//	 counter."
//
// Firebending is the shared attack trigger (CR 702.189a) with X read as
// it resolves, from the player counter Aang, Airbending Master and
// Katara, Waterbending Master use (experienceCounters), so a counter
// gained in response counts. With none it adds nothing.
//
// The cast trigger is Katara's shape with "during combat" in place of
// "during an opponent's turn": the caster is Zuko's controller and the
// turn is in its combat phase (any step of it, on any player's turn). The
// counter goes on you, not on Zuko, so it stays when he leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33db6dc8-2c88-4a7f-8d4d-3bc9fb6544b2",
		Name:            "Zuko, Firebending Master",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zuko, Firebending Master — firebending X, where X is your experience counters",
				func(g *game.Game, item *game.StackItem) error {
					// None adds nothing but still firebends (#5276).
					return firebending(experienceCounters(g, item.Controller)).Apply(NewContext(g, item))
				}),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && game.PhaseOf(g.Turn.Step) == game.PhaseCombat
			}, "Zuko, Firebending Master — you get an experience counter", youGetAnExperienceCounter),
		},
	})
}
