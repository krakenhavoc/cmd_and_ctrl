package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quest for the Goblin Lord — Enchantment {R} (Worldwake, S82 OLD b487):
//
//	"Whenever a Goblin you control enters, you may put a quest counter on this
//	 enchantment.
//	 As long as this enchantment has five or more quest counters on it, creatures
//	 you control get +2/+0."
//
// One optional trigger per Goblin that enters under your control (the "you may" is
// asked as it resolves, CR 603.5; "Goblin" is read through the enchantment's text, so
// a changeling counts). The anthem is a layer 7c static whose AppliesTo reads the
// enchantment's own quest counters, so the fifth counter switches it on at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e19baba-2d66-4e3a-b2bf-cb27dce89f44",
		Name:         "Quest for the Goblin Lord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Goblin"))
			}, "Quest for the Goblin Lord — put a quest counter", effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})),
				"Quest for the Goblin Lord — put a quest counter on it?"),
		},
		Static: []game.StaticAbility{
			effects.S82OldB487QuestAnthem("Creatures you control get +2/+0 while this has five or more quest counters.", 5, 2, 0),
		},
	})
}
