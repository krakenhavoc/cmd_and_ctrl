package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnus the Red — Legendary Creature — Demon Primarch {3}{U}{R}, 4/5
// (Warhammer 40,000 Commander, S82 GAP batch 194, #5948):
//
//	"Flying
//	 Unearthly Power — Instant and sorcery spells you cast cost {1} less to cast
//	 for each creature token you control.
//	 Blade of Magnus — Whenever Magnus the Red deals combat damage to a player,
//	 create a 3/3 red Spawn creature token."
//
// The reduction is counted as each spell is priced (CR 601.2f) and spends
// against generic mana only. The token is the printed 3/3 red Spawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86c9b01e-18a2-4616-8e04-5343f97d2efe",
		Name:            "Magnus the Red",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.GapB194CreatureTokensYouControl,
				"Instant and sorcery spells you cast cost {1} less to cast for each creature token you control.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Magnus the Red — create a 3/3 red Spawn creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 red Spawn"), N: 1})),
		},
	})
}
