package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, Seeking Revenge — Legendary Creature — Human Warrior Ally
// {3}{U/B}, 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"As an additional cost to cast this spell, you may waterbend {2}.
//	 (While paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}.)
//	 When Katara enters, draw a card, then discard a card unless her
//	 additional cost was paid.
//	 Katara gets +1/+1 for each Lesson card in your graveyard."
//
// The waterbend is an optional additional cost (OptionalWaterbend,
// CR 701.67a), Spirit Water Revival's shape. Whether it was paid rides
// the permanent's provenance (CR 400.7d, Offspring's read) and is copied
// onto the trigger as it triggers, so the answer is the cast's even if
// Katara has left by the time it resolves. Paid, the trigger only draws;
// unpaid, it draws and then discards once the draw has resolved
// (drawThenDiscard, CR 608.2c). A Katara that wasn't cast (reanimated,
// flickered) had no additional cost paid, so she loots.
//
// The pump is a layer 7c modify (CR 613.4c) counting the Lesson cards in
// your graveyard on every recompute (avB4LessonsInGraveyard).
//
// No simplification.
func init() {
	const etbLabel = "Katara, Seeking Revenge — draw a card, then discard a card unless her additional cost was paid"
	Register(Spec{
		OracleID:      "124cc815-0877-40f2-98c8-e7c4f8cd6598",
		Name:          "Katara, Seeking Revenge",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 1},
		OptionalCosts: []game.AdditionalCost{OptionalWaterbend("{2}")},
		Static: []game.StaticAbility{{
			Label:     "Katara gets +1/+1 for each Lesson card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := avB4LessonsInGraveyard(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       etbLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, etbLabel)
				if game.OptionalCostTimesPaid(*source, source.Provenance.OptionalCosts, game.WaterbendKey) > 0 {
					item.Params.Amount = 1
				}
				return item
			},
			Effect: kataraSeekingRevengeLoot,
		}},
	})
}

// kataraSeekingRevengeLoot is the enters trigger's resolution:
// Params.Amount is 1 when the waterbend was paid.
func kataraSeekingRevengeLoot(g *game.Game, item *game.StackItem) error {
	if item.Params.Amount > 0 {
		return DrawCards{N: 1}.Apply(NewContext(g, item))
	}
	return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
}
