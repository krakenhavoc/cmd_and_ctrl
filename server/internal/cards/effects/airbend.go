package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// airbend.go — S22: the exile-a-target-permanent-with-permission
// primitive, and the AVATAR keyword action built on it.
//
// S21 sub-PR 6 (#232) gave exile a permission that names a player;
// its primitive picks cards off the top of a library. Airbend needs
// the other shape — a card someone TARGETED, leaving the
// battlefield, with the permission granted to its OWNER rather than
// to the player who exiled it. The permission type already allowed
// that (Player is a free field); nothing could produce it.
//
// The two halves of airbend that were missing from the permission
// itself — no expiry, and a cost paid instead of the printed one —
// live on game.ExilePlayPermission as WhileExiled and CostOverride.
// See server/internal/game/exile_play.go for both, including the
// declared simplification on CostOverride.

// AirbendCost is the price the owner of an airbent card pays
// instead of its mana cost. Named rather than inlined because every
// airbend card prints the same reminder text and the same number,
// and a typo'd brace string in one card file would be a silently
// free spell.
const AirbendCost = "{2}"

// ExileWithPermission exiles one specific card — normally a
// targeted battlefield permanent — and grants someone permission to
// play it from exile.
//
// Sibling of ExileTopWithPermission, which is the library-facing
// shape. The difference that matters is who gets the grant: impulse
// exile hands the card to the player who exiled it, and airbend
// hands it back to its owner. GrantTo left zero means "the owner",
// which is the airbend case and is NOT the same as "the controller
// of the effect".
//
// WhileExiled makes the grant unbounded ("while it's exiled");
// leaving it false gives the impulse-exile default of "until end of
// this turn".
type ExileWithPermission struct {
	// Target is the card to exile. Any zone — the primitive routes
	// through the ordinary exile path, so a battlefield permanent
	// gets its LKI snapshot and its LTB trigger.
	Target uuid.UUID

	// GrantTo is who may play it. uuid.Nil means the card's owner.
	GrantTo uuid.UUID

	// CastOnly restricts the grant to casting, stranding a land.
	CastOnly bool

	// AnyColor lets the holder spend mana as though it were mana of
	// any color.
	AnyColor bool

	// WhileExiled makes the grant last as long as the card stays in
	// exile rather than expiring at end of turn.
	WhileExiled bool

	// CostOverride is a mana cost paid instead of the card's printed
	// one, in Scryfall brace notation. Empty means the printed cost.
	CostOverride string
}

func (e ExileWithPermission) Apply(ctx *Context) error {
	return ctx.Game.ExileCardWithPermissionForEffect(e.Target, game.ExilePlayPermission{
		Player:       e.GrantTo,
		CastOnly:     e.CastOnly,
		AnyColor:     e.AnyColor,
		WhileExiled:  e.WhileExiled,
		CostOverride: e.CostOverride,
	})
}

// Airbend is the AVATAR keyword action: "Exile it. While it's
// exiled, its owner may cast it for {2} rather than its mana cost."
//
// CastOnly is set because the reminder text says CAST (CR 305.1 —
// playing a land is not casting), so an airbent land would be
// stranded in exile. Every airbend card in the set targets a
// nonland permanent or a creature, so the flag is a statement of
// intent rather than a live restriction; it is set anyway, because
// the alternative is a primitive that quietly hands out land drops
// the moment a card widens its target clause.
//
// The grant goes to the card's owner (GrantTo left zero), which is
// what makes airbend removal-with-a-refund rather than theft: you
// exile an opponent's blocker and they get it back cheaply. That is
// the printed card, and getting the direction wrong would turn a
// tempo play into a Mind Control.
type Airbend struct {
	Target uuid.UUID
}

func (a Airbend) Apply(ctx *Context) error {
	return ExileWithPermission{
		Target:       a.Target,
		CastOnly:     true,
		WhileExiled:  true,
		CostOverride: AirbendCost,
	}.Apply(ctx)
}
