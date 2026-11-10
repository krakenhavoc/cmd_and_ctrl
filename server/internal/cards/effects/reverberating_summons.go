package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reverberating Summons — Enchantment {1}{R} (Tarkir: Dragonstorm,
// #4463):
//
//	"At the beginning of each combat, if you've cast two or more spells
//	 this turn, this enchantment becomes a 3/3 Monk creature with haste in
//	 addition to its other types until end of turn.
//	 {1}{R}, Discard your hand, Sacrifice this enchantment: Draw two
//	 cards."
//
// The trigger is Gau's shape: it watches the beginning of every combat (each
// player's, the Summons' controller's two spells counting even on an
// opponent's turn) and holds the intervening "if" (CR 603.4) both as the step
// begins and as the trigger resolves, read off the controller's per-turn cast
// tally (CastTallyFor, a copy is not cast). The animation is Reverberating
// Summons' own object only: a layer 4 Creature type with the Monk subtype in
// addition to Enchantment, a layer 7b 3/3 and haste, all until end of turn
// (BecomeCreatureUntilEOT); one that has left or become a new object is not
// animated (CR 400.7). It is a creature only until cleanup, so the sacrifice
// ability works the same at any time. The ability's costs are Connecting the
// Dots': mana, the whole hand, and the enchantment itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fd78bf7-43ba-4e09-834b-14bfc1fd9d0d",
		Name:         "Reverberating Summons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventStepBegan},
			AppliesTo: AllOf(
				StepBegan(game.StepBeginCombat, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.CastTallyFor(source.Controller).Total >= 2
				},
			),
			Key: "Reverberating Summons — becomes a 3/3 Monk creature with haste until end of turn",
			Effect: func(g *game.Game, item *game.StackItem) error {
				// CR 603.4: the intervening if again on resolution.
				if g.CastTallyFor(item.Controller).Total < 2 || !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (BecomeCreatureUntilEOT{
					Target:       item.SourceCardID,
					Types:        []string{"Creature"},
					Subtypes:     []string{"Monk"},
					SetPower:     3,
					SetToughness: 3,
					Label:        "Reverberating Summons — becomes a 3/3 Monk creature",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Label:    "Reverberating Summons — haste until end of turn",
				}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Discard your hand, Sacrifice this enchantment: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{R}"), DiscardYourHand(), SacrificeThis()),
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
