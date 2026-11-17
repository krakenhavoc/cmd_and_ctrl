package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heliod, the Radiant Dawn // Heliod, the Warped Eclipse — {2}{W}{W}
// Legendary Enchantment Creature — God 4/4 // Legendary Enchantment
// Creature — Phyrexian God 4/6 (March of the Machine):
//
//	Front: "When Heliod enters, return target enchantment card that isn't a
//	        God from your graveyard to your hand.
//	        {3}{U/P}: Transform Heliod. Activate only as a sorcery."
//	Back:  "You may cast spells as though they had flash.
//	        Spells you cast cost {1} less to cast for each card your
//	        opponents have drawn this turn."
//
// The reduction counts every card an opponent of Heliod's controller has
// drawn this turn (Game.DrawnThisTurn, appended only by an actual draw and
// never pruned, so a card drawn and then discarded still counts, and a
// card put into a hand without being drawn does not). It lowers generic
// mana only, as every cost reduction does (CR 601.2f).
//
// No simplification.
func init() {
	const oracle = "c27c6ba9-8d72-4101-b3a5-1e8d81755568"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Heliod, the Radiant Dawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Heliod, the Radiant Dawn — return target enchantment card that isn't a God from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
				}),
				TargetCardInGraveyard("target enchantment card that isn't a God from your graveyard", YouOwn(), Enchantment(), Not(HasSubtype("God")))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{U/P}: Transform Heliod. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{U/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Heliod, the Warped Eclipse",
		Completeness: CompletenessFull,
		CastTimings:  []game.CastTimingRule{CastAsThoughFlash()},
		CostModifiers: []game.CostModifier{
			CostsLessEach(s82Mom16CardsOpponentsDrewThisTurn,
				"Spells you cast cost {1} less to cast for each card your opponents have drawn this turn.",
				YourSpell()),
		},
	})
}
