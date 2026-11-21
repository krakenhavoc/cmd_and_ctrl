package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_prowess_test.go — ADR 0018's #3013 amendment,
// sub-PR 3, closing #1530 item 1. Prowess has a declared footprint (it
// reads and pumps its own source and emits nothing), so a Monastery
// Mentor board and a Sokka, Tenacious Tactician board, each prowess
// plus a token trigger, go on the stack with no prompt under "Ask only
// when the order matters". The batch still asks when the order can
// show: something watches the token enter, a trigger the check cannot
// read, or a token whose own keyword trigger watches what the batch
// does.

const (
	EvolveTokenProbe = "test-3013-evolve-token-probe"
	MonkTokenProbe   = "test-3013-monk-token-probe"
)

func init() {
	// "Whenever you cast a noncreature spell, create a 1/1 token with
	// evolve." No catalogued token has evolve; the shape is what
	// tokenTemplateInert used to miss.
	effects.Register(effects.Spec{
		OracleID: EvolveTokenProbe,
		Name:     "Evolve Token Probe",
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Evolve Token Probe — create a 1/1 with evolve",
				effects.Do(effects.CreateToken{Template: game.Card{
					Name: "Evolver", TypeLine: "Token Creature — Ooze", Power: 1, Toughness: 1,
					Keywords: []string{game.KeywordEvolve},
				}, N: 1})),
		},
	})
	// The same with Monastery Mentor's own Monk, whose prowess watches a
	// cast: nothing in a batch casts a spell.
	effects.Register(effects.Spec{
		OracleID: MonkTokenProbe,
		Name:     "Monk Token Probe",
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Monk Token Probe — create a 1/1 Monk with prowess",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Monk with prowess"), N: 1})),
		},
	})
}
