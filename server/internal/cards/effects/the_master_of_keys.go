package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Master of Keys — Legendary Enchantment Creature — Horror
// {X}{W}{U}{B}, 3/3 (Duskmourn Commander, #5368):
//
//	"Flying
//	 When The Master of Keys enters, put X +1/+1 counters on it and mill
//	 twice X cards.
//	 Each enchantment card in your graveyard has escape. The escape cost is
//	 equal to the card's mana cost plus exile three other cards from your
//	 graveyard."
//
// The enters trigger reads the X the spell was cast with off the
// permanent as it is built (Ajani's Anguish's etbDealsCastXDamage) and
// carries it on the item, so it is exact even if the creature has left by
// resolution; a copy or a cheat onto the battlefield has X = 0 and does
// nothing. The counters go on first, then the mill (twice X).
//
// The escape line is Underworld Breach's standing permission, narrowed to
// enchantment CARDS in the graveyard (the card, not the spell; CR 702.138):
// the escape cost is the card's own mana cost plus exiling three other
// cards, and the permission is gone the instant this leaves the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37dc566e-a91a-448c-9967-b178a93ffb62",
		Name:            "The Master of Keys",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			Filter:                  game.PermissionFilter{CardInZone: true, CardType: "enchantment"},
			AltCostKey:              "escape",
			ExileOtherFromGraveyard: 3,
			Label:                   "Escape — its mana cost, Exile three other cards from your graveyard",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "The Master of Keys — put X +1/+1 counters on it and mill twice X cards",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "The Master of Keys — put X +1/+1 counters on it and mill twice X cards")
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
					return err
				}
				return MillCards{N: 2 * x}.Apply(ctx)
			},
		}},
	})
}
