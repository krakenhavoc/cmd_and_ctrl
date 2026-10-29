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

// bilboCastFromGraveyard offers the one cast over the artifact, instant
// and sorcery cards in the controller's graveyard. A land is never cast,
// so an artifact land is not offered.
func bilboCastFromGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	cards := graveyardCardIDs(ctx, you, func(c game.Card) bool {
		return !c.IsLand() && (c.IsArtifact() || c.IsInstant() || c.IsSorcery())
	})
	if len(cards) == 0 || you == uuid.Nil {
		return nil
	}
	perm := game.CastPermission{
		Player:                           you,
		Zone:                             game.ZoneGraveyard,
		Duration:                         g.UntilEndOfTurnDuration(),
		AltCostKey:                       game.AltCostKeyGrantedCost,
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
