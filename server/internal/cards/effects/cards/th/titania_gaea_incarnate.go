package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titania, Gaea Incarnate — Legendary Creature — Elemental Avatar, */*. The
// combined back face of Titania, Voice of Gaea and Argoth, Sanctum of
// Nature (CR 712.5b); never a card in a deck, only the melded permanent the
// two become (ADR 0145, #2699), here for The Brothers' War batch 20
// (#5482):
//
//	"Reach, vigilance, trample, haste
//	 Titania's power and toughness are each equal to the number of lands
//	 you control.
//	 When Titania enters, return all land cards from your graveyard to the
//	 battlefield tapped.
//	 {3}{G}: Put four +1/+1 counters on target land you control. It becomes
//	 a 0/0 Elemental creature with haste. It's still a land."
//
// The keywords are the back face's printed ones, carried by the melded
// permanent from the deck import. The size is a characteristic-defining
// ability over the lands (landsYouControlPT). The enters trigger reads the
// graveyard once, before anything moves, and the lands enter together
// tapped. The activated ability is Cosmium Confluence's animation over one
// targeted land: the counters go on first, then it becomes a 0/0 Elemental
// creature with haste that is still a land, with no duration.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43230d70-4d60-4ceb-bb45-fbc043280652",
		Name:         "Titania, Gaea Incarnate",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Titania's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Titania, Gaea Incarnate — return all land cards from your graveyard to the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B10ReturnAllLandCardsFromGraveyardTapped(effects.NewContext(g, item), item.Controller)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Put four +1/+1 counters on target land you control. It becomes a 0/0 Elemental creature with haste. It's still a land.",
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetPermanent("target land you control", effects.And(effects.Land(), effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.LegalTargetCards(item, g) {
					return effects.AnimateLandWithCounters(ctx, id, 4, "Elemental")
				}
				return nil
			},
		}},
	})
}
