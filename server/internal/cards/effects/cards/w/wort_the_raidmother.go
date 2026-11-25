package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wort, the Raidmother — {4}{R/G}{R/G} Legendary Creature — Goblin Shaman 3/3
// (Commander 2020):
//
//	"When Wort enters, create two 1/1 red and green Goblin Warrior creature
//	 tokens.
//	 Each red or green instant or sorcery spell you cast has conspire. (As you
//	 cast the spell, you may tap two untapped creatures you control that share
//	 a color with it. When you do, copy it and you may choose new targets for
//	 the copy.)"
//
// The tokens are a plain entry trigger. The grant is Raiding Schemes' conspire
// (ConspireForSpellsYouCast, one conspire instance per source, CR 702.78b)
// over instant and sorcery spells that are red or green.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6b67c7e-3b24-4f60-8e3a-c6ca45229d7f",
		Name:         "Wort, the Raidmother",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wort, the Raidmother — create two 1/1 red and green Goblin Warrior creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red and green Goblin Warrior"), N: 2})),
		},
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			effects.ConspireForSpellsYouCast("Each red or green instant or sorcery spell you cast has conspire.",
				effects.And(effects.Or(effects.Instant(), effects.Sorcery()), effects.Or(effects.OfColor("R"), effects.OfColor("G")))),
		},
	})
}
