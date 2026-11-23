package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Mercadia — Battle — Siege, defense 4, for {1}{R}, and its
// back face Kyren Flamewright — Creature — Goblin Spellshaper, 3/3
// (March of the Machine):
//
//	"When this Siege enters, you may discard a card. If you do, draw two
//	 cards."
//
//	"{2}{R}, {T}, Discard a card: Create two 1/1 blue and red Elemental
//	 creature tokens. Creatures you control get +1/+0 and gain haste
//	 until end of turn."
//
// The defense counters, the protector prompt and the defeat are engine
// behaviour keyed on the battle type (Invasion of Karsus' note). The
// loot is Bitter Reunion's. The back face registers under
// "<oracle_id>#1". Its pump and haste fix the creatures as the ability
// resolves, so they include the two new tokens (CR 611.2c).
//
// No simplification.
const InvasionOfMercadiaOracleID = "dd5a9ea5-0f27-4492-97d1-1a9970094c19"

func init() {
	const loot = "Invasion of Mercadia — you may discard a card; if you do, draw two cards"
	effects.Register(effects.Spec{
		OracleID:     InvasionOfMercadiaOracleID,
		Name:         "Invasion of Mercadia",
		Completeness: effects.CompletenessFull,
		Battle: &effects.BattleSpec{
			Defense: 4,
			Subtype: effects.BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(loot, func(g *game.Game, item *game.StackItem) error {
				return effects.MayDiscardACardToDrawTwo(loot)(effects.NewContext(g, item))
			}),
			effects.DefeatedTrigger("Invasion of Mercadia — defeated: exile it, then cast Kyren Flamewright", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     InvasionOfMercadiaOracleID + "#1",
		Name:         "Kyren Flamewright",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: Create two 1/1 blue and red Elemental creature tokens. Creatures you control get +1/+0 and gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("1/1 blue and red Elemental")), N: 2}).Apply(ctx); err != nil {
					return err
				}
				mine := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Label: "Kyren Flamewright — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"haste"}, Label: "Kyren Flamewright — haste"}.Apply(ctx)
			},
		}},
	})
}
