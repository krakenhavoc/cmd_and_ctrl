package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Six — Legendary Creature — Treefolk {2}{G}, 2/4:
//
//	"Reach
//	 Whenever Six attacks, mill three cards. You may put a land card
//	 from among them into your hand.
//	 During your turn, nonland permanent cards in your graveyard have
//	 retrace."
//
// The first two lines are the card that gets attacked with: a
// three-mana 2/4 reach body that turns every attack into a land and
// three cards of graveyard fuel. Reach rides PrintedKeywords; the
// attack trigger is the shared mill-then-take shape
// (mill_then_take.go), filtered to lands.
//
// The mill is NOT optional — "mill three cards" is flat — and only
// the pick is a "may", which is why the trigger has no
// OptionalPrompt and the choose-cards prompt floors at zero.
//
// The third line is a STANDING cast permission (ADR 0066, #2528),
// declared on Spec.CastPermissions and derived from the battlefield on
// every query, never stored — so it lasts exactly as long as Six does,
// covers a card milled by this very attack, and two Sixes compose. Its
// filter is "nonland permanent cards" (NonLandPermanentOnly: an artifact,
// creature, enchantment, planeswalker or battle, never a land and never
// an instant or sorcery), its zone is the graveyard, and its timing is
// TimingYourTurnOnly — "during your turn", with the card's own timing
// still in force on it, so a permanent spell is a main-phase cast and a
// flash creature is castable at instant speed, on Six's controller's
// turn only.
//
// Retrace is not an alternative cost: the card is cast for its PRINTED
// mana cost and a land card is discarded in addition (CR 702.81a). The
// permission synthesises a priced offer of that shape (AltCostKey
// "retrace", DiscardLandCard); see AlternativeCost.DiscardFromHand for
// why an additional cost rides an offer. Claiming the offer is what opens
// the graveyard, so the discard is owed on exactly the graveyard cast and
// never on the hand cast of the same card. A card that prints its own
// graveyard cast (flashback, retrace) keeps its own price, as under every
// standing grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbcbdf37-c40f-4068-b4a7-a849cab1056c",
		Name:         "Six",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:            game.ZoneGraveyard,
			Filter:          game.PermissionFilter{NonLandPermanentOnly: true},
			AltCostKey:      game.AltCostKeyRetrace,
			DiscardLandCard: true,
			Timing:          game.TimingYourTurnOnly,
			Label:           "Retrace — discard a land card (Six)",
		}},
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Six — mill three cards", sixAttackMill),
		},
	})
}

// sixAttackMill is the attack trigger: mill three, then offer a land
// from among them.
func sixAttackMill(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 3,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return mayTakeOneFromAmongThem(ctx, milled, Land(),
				"Six — you may put a land card from among them into your hand")
		},
	}.Apply(NewContext(g, item))
}
