package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lumra, Bellow of the Woods — Legendary Creature — Elemental Bear
// {4}{G}{G}, */* (EDHREC rank 1129):
//
//	"Reach, vigilance
//	 Lumra's power and toughness are each equal to the number of lands
//	 you control.
//	 When Lumra enters, mill four cards. Then return all land cards
//	 from your graveyard to the battlefield tapped."
//
// The lands deck's six-drop. Reach and vigilance ride PrintedKeywords;
// the P/T is a Layer 7a characteristic-defining ability (Adeline's
// shape) that SETS both values to the land count on every recompute,
// so the lands the ETB returns are counted the moment they land. The
// ETB is the mill then Splendid Reclamation's body — every land card
// in the graveyard, the four just milled included, back at once: one
// entry, tapped as they enter (#1867).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97a84e9d-bfc4-4ca2-b1e8-908dba56ccdb",
		Name:            "Lumra, Bellow of the Woods",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b10LandsControlled(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lumra — mill four, then return all land cards from your graveyard tapped", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (MillCards{Player: item.Controller, N: 4}).Apply(ctx); err != nil {
					return err
				}
				return b10ReturnAllLandCardsFromGraveyardTapped(ctx, item.Controller)
			}),
		},
	})
}
