package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bilbo, Thief in the Night — Legendary Creature — Halfling Rogue {1}{U},
// 2/2 (The Hobbit):
//
//	"Spells you cast from anywhere other than your hand cost {1} less to
//	 cast.
//	 Whenever Bilbo attacks, you may cast an artifact, instant, or sorcery
//	 spell from your graveyard. If an instant or sorcery spell cast this
//	 way would be put into your graveyard, exile it instead."
//
// The discount is Fortune Teller's Talent's cost modifier keyed on the
// zone the cast comes FROM (castFromOutsideYourHand over
// CostQuery.FromZone): a graveyard, exile, library or command-zone cast is
// {1} cheaper, a hand cast is not, and it prices the attack trigger's own
// cast too.
//
// The cast happens while the attack trigger resolves (CR 608.2g), through
// the resolution_cast prompt (#4250, game.OfferResolutionCastForEffect),
// the door Torrential Gearhulk and Emet-Selch of the Third Seat use: ONE
// pass-closed, any-timing, cast-only permission over every artifact,
// instant and sorcery card in the graveyard, at the card's own mana cost.
// The first cast spends it, which is "cast a spell" (one); the decline is
// the "you may", and leaves the cards where they are. "If an instant or
// sorcery spell cast this way would be put into your graveyard, exile it
// instead" is the permission's ExileInsteadOfGraveyard narrowed by
// ExileInsteadOnlyInstantOrSorcery (#4505): an artifact spell cast this
// way that is countered goes to the graveyard, as printed, and an instant
// or sorcery returned to a hand or a library goes there.
//
// "An artifact, instant, or sorcery SPELL" is judged on the spell, the
// face being cast (#4509: CR 601.3e, 709.3a, 712.11c, 715.3a), not on the
// card in the graveyard. So the permission carries the filter
// (ArtifactInstantOrSorceryOnly, about the spell) and the offer is over
// every card with a face it opens: a creature card whose Adventure is an
// instant or sorcery is offered and only its Adventure may be cast (the
// card's ruling of 2026-06-29), a modal DFC is offered for an artifact,
// instant or sorcery back face (Tergrid's Lantern, Deadly Vanity) and
// cast only as that face, and a Town whose Adventure is a sorcery is
// offered for its Adventure though the card is a land. An Adventure cast
// this way that resolves may go on its adventure or be exiled by
// Bilbo's replacement, its controller's choice (CR 616.1, the same
// ruling); one that is countered is exiled by Bilbo's.
//
// No simplification. The Hobbit (#4454), seam #4505.
const (
	bilboThiefOracleID = "72a9a85d-7cfc-4b3c-82db-9d35be6c0982"
	bilboThiefName     = "Bilbo, Thief in the Night"
)

func init() {
	Register(Spec{
		OracleID:     bilboThiefOracleID,
		Name:         bilboThiefName,
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast from anywhere other than your hand cost {1} less to cast.",
				YourSpell(), castFromOutsideYourHand()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(bilboThiefName+" — you may cast an artifact, instant, or sorcery spell from your graveyard",
				bilboCastFromGraveyard),
		},
	})
}

// bilboCastFromGraveyard offers the one cast over the cards in the
// controller's graveyard with an artifact, instant or sorcery spell to
// cast, face by face (CastPermission.OffersASpell). A land is never
// cast, so an artifact land is not offered.
func bilboCastFromGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	if you == uuid.Nil {
		return nil
	}
	perm := game.CastPermission{
		Player:     you,
		Zone:       game.ZoneGraveyard,
		Duration:   g.UntilEndOfTurnDuration(),
		AltCostKey: game.AltCostKeyGrantedCost,
		// About the spell, so judged on the face being cast (#4509).
		Filter:                           game.PermissionFilter{ArtifactInstantOrSorceryOnly: true},
		Timing:                           game.TimingFlash,
		CastOnly:                         true,
		ExileInsteadOfGraveyard:          true,
		ExileInsteadOnlyInstantOrSorcery: true,
		// A pass closes the window; nothing moves, the cards stay in the
		// graveyard.
		LapseOnPass: game.LapseStaysInHand,
		Source:      ctx.Source(),
		SourceName:  bilboThiefName,
		Label:       bilboThiefName + " — cast an artifact, instant, or sorcery spell from your graveyard",
	}
	cards := graveyardCardIDs(ctx, you, perm.OffersASpell)
	if len(cards) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       bilboThiefName,
		Cards:        cards,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it",
		DeclineLabel: declineLeaveInGraveyard,
	})
	return nil
}
