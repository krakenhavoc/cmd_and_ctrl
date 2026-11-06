package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambling Stormshell — Creature — Turtle {3}{U}{U}, 5/9 (Tarkir:
// Dragonstorm, #4463):
//
//	"Ward {2}
//	 Whenever this creature attacks, put three stun counters on it and
//	 draw three cards. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)
//	 Whenever you cast a Turtle spell, untap this creature."
//
// Ward is the shared trigger. The attack trigger puts the counters on the
// Stormshell only while it is still on the battlefield, and draws the
// three cards either way. Stun counters are the engine's untap
// replacement (CR 122.1d), applied wherever a permanent would untap, so
// the Turtle trigger's untap removes a stun counter instead while one is
// there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75070eaf-c931-46e5-9aa1-b849e33580c6",
		Name:         "Ambling Stormshell",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Ambling Stormshell — ward {2}"),
			WheneverThisAttacks("Ambling Stormshell — put three stun counters on it and draw three cards",
				Do(CounterOnThis{Kind: game.CounterStun, N: 3}, DrawCards{N: 3})),
			WheneverYouCast(Subtype("Turtle"), "Ambling Stormshell — untap this creature",
				func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
