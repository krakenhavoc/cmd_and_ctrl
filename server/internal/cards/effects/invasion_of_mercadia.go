package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const invasionOfMercadiaOracleID = "dd5a9ea5-0f27-4492-97d1-1a9970094c19"

func init() {
	const loot = "Invasion of Mercadia — you may discard a card; if you do, draw two cards"
	Register(Spec{
		OracleID:     invasionOfMercadiaOracleID,
		Name:         "Invasion of Mercadia",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(loot, func(g *game.Game, item *game.StackItem) error {
				return mayDiscardACardToDrawTwo(loot)(NewContext(g, item))
			}),
			DefeatedTrigger("Invasion of Mercadia — defeated: exile it, then cast Kyren Flamewright", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:     invasionOfMercadiaOracleID + "#1",
		Name:         "Kyren Flamewright",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: Create two 1/1 blue and red Elemental creature tokens. Creatures you control get +1/+0 and gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CreateToken{Template: ctx.TextToken(TokenCard("1/1 blue and red Elemental")), N: 2}).Apply(ctx); err != nil {
					return err
				}
				mine := And(Creature(), YouControl())
				if err := (BoostUntilEOT{Match: mine, Power: 1, Label: "Kyren Flamewright — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"haste"}, Label: "Kyren Flamewright — haste"}.Apply(ctx)
			},
		}},
	})
}
