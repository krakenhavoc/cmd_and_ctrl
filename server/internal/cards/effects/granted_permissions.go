package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_permissions.go — S42, ADR 0066: the card-side vocabulary for
// "an effect lets you cast a card from somewhere its own text does not
// allow".
//
// Two shapes, and they are the two scopes of game.CastPermission:
//
//   - A permission an EFFECT grants, over cards named when it
//     resolves. Snapcaster Mage names one card; Past in Flames and The
//     Grim Captain's Locker name every matching card in the graveyard.
//     CR 611.2c locks that set at resolution, so both are one call to
//     GrantCastFromYourGraveyard and a card that reaches the graveyard
//     a moment later has nothing.
//   - A permission a PERMANENT grants while it is on the battlefield.
//     That is not code at all: it is `Spec.CastPermissions`, derived
//     from the battlefield on every query (Underworld Breach, Bolas's
//     Citadel, the Future Sight family).
//
// The price is spelled the same way in both, because a granted
// permission synthesises the alternative cost the cast is judged
// under: `AltCostKey` shares the keyword's key, so CR 702.34a's
// "if the flashback cost was paid, exile it" and CR 702.138b's
// "escaped" work however the permission arrived.

// GrantFlashbackToCard is Snapcaster Mage's clause: "target instant or
// sorcery card in your graveyard gains flashback until end of turn.
// The flashback cost is equal to its mana cost" (CR 702.34).
//
// The cost is left empty, which the permission reads as "its mana
// cost" — the printed wording — and ExileOnResolution is what makes it
// flashback rather than a free Regrowth: CR 702.34a exiles the card
// instead of putting it anywhere else any time it would leave the
// stack, so a Snapcaster'd Brainstorm is cast once and gone even if it
// is countered.
type GrantFlashbackToCard struct {
	// Target is the card in a graveyard that gains flashback.
	Target uuid.UUID
	// Label is the clause as printed, for the client's cost picker.
	Label string
	// Cost overrides the flashback cost when it is not the card's mana
	// cost — "gains flashback {0}" (Archmage's Newt, saddled). Empty
	// keeps "equal to its mana cost".
	Cost string
}

func (e GrantFlashbackToCard) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	label := e.Label
	if label == "" {
		label = "Flashback — its mana cost"
	}
	ctx.Game.GrantCastPermissionOverCardForEffect(e.Target, game.CastPermission{
		Player:            ctx.Controller(),
		Zone:              game.ZoneGraveyard,
		AltCostKey:        "flashback",
		Cost:              e.Cost,
		ExileOnResolution: true,
		// Duration left zero: the one write path reads that as "until
		// end of turn" and stamps it against this turn, which is what
		// every card here prints.
		Source: ctx.Source(),
		Label:  label,
	})
	return nil
}

// GrantCastFromYourGraveyard is the set-at-resolution shape: "each
// <kind> card in your graveyard gains <keyword> until end of turn".
//
// CR 611.2c is the whole of it. A one-shot continuous effect locks the
// set of objects it affects when it resolves, so this walks the
// graveyard ONCE and writes the instances down. A card milled,
// discarded or killed afterwards is not in the set and gains nothing —
// which is the difference between Past in Flames and Underworld
// Breach, and the reason the second is a Spec slot rather than this.
type GrantCastFromYourGraveyard struct {
	// Filter narrows which cards in the graveyard qualify.
	Filter game.PermissionFilter

	// AltCostKey is the keyword the granted cost is claimed under —
	// "flashback", "escape". Shared with the printed keywords on
	// purpose (ADR 0066 decision 3).
	AltCostKey string

	// Cost is the mana cost the granted keyword charges. Empty means
	// "equal to its mana cost", which is what Past in Flames and
	// Snapcaster both print.
	Cost string

	// ExileOtherFromGraveyard is escape's "exile N other cards from
	// your graveyard" (CR 702.138a) — four for The Grim Captain's
	// Locker.
	ExileOtherFromGraveyard int

	// ExileOnResolution is flashback's CR 702.34a clause. Escape does
	// NOT set it: an escaped card goes to the battlefield or the
	// graveyard like any other and escapes again next time.
	ExileOnResolution bool

	// Label is the clause as printed.
	Label string
}

func (e GrantCastFromYourGraveyard) Apply(ctx *Context) error {
	controller := ctx.Controller()
	p := ctx.PlayerByID(controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var cards []game.Card
	for _, c := range p.Graveyard.Cards {
		if e.Filter.Matches(c) {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:                  controller,
		Zone:                    game.ZoneGraveyard,
		AltCostKey:              e.AltCostKey,
		Cost:                    e.Cost,
		ExileOtherFromGraveyard: e.ExileOtherFromGraveyard,
		ExileOnResolution:       e.ExileOnResolution,
		Source:                  ctx.Source(),
		Label:                   e.Label,
	}, cards)
	return nil
}

// PlayFromTopOfYourLibrary is the Spec-slot constructor for the Future
// Sight family: "you may play <what> from the top of your library"
// (CR 401.5).
//
// A card that declares one MUST also declare Spec.LibraryTopVisible —
// a card you cannot see is a card you cannot play, and every printed
// card in the family carries both halves. The engine refuses the cast
// if the visibility is missing, so the failure is a card that does
// nothing rather than one that cheats.
func PlayFromTopOfYourLibrary(filter game.PermissionFilter, label string) game.CastPermission {
	return game.CastPermission{
		Zone:             game.ZoneLibrary,
		TopOfLibraryOnly: true,
		Filter:           filter,
		Label:            label,
	}
}

// OfferGraveyardCast is "you may cast target <card> from your graveyard.
// If that spell would be put into your graveyard, exile it instead"
// (Emet-Selch of the Third Seat, #3785): a cast made as the ability
// resolves (CR 608.2g), at the card's mana cost, whose spell is exiled
// if it would be put into its owner's graveyard — as it resolves, or
// when it is countered or fizzles — and goes anywhere else it is sent
// (game.CastPermission.ExileInsteadOfGraveyard, #3863).
//
// It is Torrential Gearhulk's grant with the price left at "its mana
// cost" and a FollowUp. The offer is flash-timed and made during the
// resolution (#4250). The holder's next priority pass closes it only as the
// fallback (LapseStaysInHand: an uncast card simply stays in the graveyard), and it is claimed under the shared
// "granted-cost" key rather than "flashback", so a card that prints
// flashback is still cast for its MANA cost here (a flashback claim
// would take the card's own flashback price), and the cast counts as
// this permission's: the FollowUp runs for it.
//
// FollowUp runs only when the cast is made (game.CastPermission.FollowUp,
// #2173), which is what lets "do this only once each turn" count a
// cast and not a declined or lapsed offer. The body is handed the
// permission's Source, the ability's source object.
//
// #4250: the cast is made during the resolution, through the
// resolution_cast prompt over the card (paid, so the caster may tap mana
// for it then); the decline leaves it in the graveyard.
type OfferGraveyardCast struct {
	// Target is the card in the controller's graveyard.
	Target uuid.UUID
	// SourceName is the granting card's name, for the log.
	SourceName string
	// FollowUp is a registered game.RegisterCastFollowUp key, or empty.
	FollowUp game.CastFollowUpKey
}

func (e OfferGraveyardCast) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	g := ctx.Game
	z := g.FindCardZoneForEffect(e.Target)
	c, ok := g.LookupCardForEffect(e.Target)
	if !ok || z == nil || z.Kind != game.ZoneGraveyard || c.Owner != ctx.Controller() {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:                  ctx.Controller(),
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		AltCostKey:              game.AltCostKeyGrantedCost,
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		ExileInsteadOfGraveyard: true,
		LapseOnPass:             game.LapseStaysInHand,
		FollowUp:                e.FollowUp,
		Source:                  ctx.Source(),
		SourceName:              e.SourceName,
		Label:                   e.SourceName + " — cast " + c.Name + " from your graveyard",
	}, []game.Card{c})
	offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), e.SourceName, game.MayCastKeywordPaidCast, declineLeaveInGraveyard, c.InstanceID)
	return nil
}
