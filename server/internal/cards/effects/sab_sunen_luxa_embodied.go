package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sab-Sunen, Luxa Embodied — Legendary Creature — God {3}{G}{U}, 6/6
// (Aetherdrift, #4464):
//
//	"Reach, trample, indestructible
//	 Sab-Sunen can't attack or block unless it has an even number of
//	 counters on it. (Zero is even.)
//	 At the beginning of your first main phase, put a +1/+1 counter on
//	 Sab-Sunen. Then if it has an odd number of counters on it, draw two
//	 cards."
//
// The restriction is Bast, Panther Goddess's conditional layer-6
// CantAttack | CantBlock, with the condition read off the counters on
// Sab-Sunen of every kind. A counter is not a characteristic, so the
// layer-6 read sees the real count, and a counter being placed or removed
// already drops the layer cache. Restrictions are checked as attackers and
// blockers are declared (CR 508.1c, 509.1b).
//
// The trigger puts the counter on Sab-Sunen while it is still the object
// the trigger came from (CR 400.7), then reads its counters at that
// moment, or as it last existed on the battlefield if it has left
// (CR 608.2h), and draws two on an odd count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07e15387-f346-4fa0-b556-d1c52d4acc07",
		Name:            "Sab-Sunen, Luxa Embodied",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample", "indestructible"},
		Static: []game.StaticAbility{{
			Label: "Sab-Sunen can't attack or block unless it has an even number of counters on it.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && sabSunenCounterTotal(source.Counters)%2 != 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Sab-Sunen — put a +1/+1 counter on it, then draw two if it has an odd number of counters",
				sabSunenFirstMain),
		},
	})
}

// sabSunenFirstMain is the first-main-phase trigger's body.
func sabSunenFirstMain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// AddCounter does nothing to a Sab-Sunen that left or came back as a
	// new object.
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	info, ok := ctx.SourcePermanent()
	if !ok || sabSunenCounterTotal(info.Counters)%2 == 0 {
		return nil
	}
	return DrawCards{N: 2}.Apply(ctx)
}

// sabSunenCounterTotal is the number of counters of every kind in
// `counters`.
func sabSunenCounterTotal(counters map[string]int) int {
	n := 0
	for _, v := range counters {
		if v > 0 {
			n += v
		}
	}
	return n
}
