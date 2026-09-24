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
// itself — a window that is a zone rather than a turn, and a cost
// paid instead of the printed one — live on game.CastPermission as
// Duration.WhileInZone and Cost.
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

	// AnyType is the wider "mana of any TYPE can be spent" (Hostage
	// Taker, #1573): colorless mana pays a {C} in the cost too.
	AnyType bool

	// WhileExiled makes the grant last as long as the card stays in
	// exile rather than expiring at end of turn.
	WhileExiled bool

	// CostOverride is a mana cost paid instead of the card's printed
	// one, in Scryfall brace notation. Empty means the printed cost.
	CostOverride string
}

func (e ExileWithPermission) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(e.Target) { // #1432
		return nil
	}
	return exileAllWithPermission(ctx.Game, []uuid.UUID{e.Target}, e.permission())
}

// permission is the grant the primitive's fields describe, before the
// holder and zone are filled in per landed card.
func (e ExileWithPermission) permission() game.CastPermission {
	perm := game.CastPermission{
		Player:   e.GrantTo,
		CastOnly: e.CastOnly,
		// AnyType implies AnyColor, and saying both keeps the grant
		// any-colour for a binary that predates AnyType (#1573).
		AnyColor: e.AnyColor || e.AnyType,
		AnyType:  e.AnyType,
		Cost:     e.CostOverride,
	}
	if e.WhileExiled {
		// CR 611.2b, "while it's exiled". Left zero the permission
		// takes the engine's default of "until end of turn", which is
		// the impulse-exile window.
		perm.Duration = game.WhileInZoneDuration()
	}
	return perm
}

// exileAllWithPermission exiles `ids` as one simultaneous event and
// grants `perm` over each card that actually LANDED in exile.
//
// The grant is stamped from the continuation, not on the next line,
// and #1304 is why. An exile can pause on the CR 903.9 prompt when the
// card is a commander — every airbend aimed at an opposing commander,
// and Appa airbending your own out of a wrath. The fire-and-forget
// exile (game.ExileCardWithPermissionForEffect) returns while the
// commander is still on the battlefield waiting for its owner's
// answer; its permission scan finds nothing in exile, so a commander
// whose owner DECLINED the command zone landed in exile with no way
// to cast it back for {2}. The continuation runs after the answer, and
// only over the cards that arrived: a commander that took the command
// zone left, but not to exile (CR 400.7), so it gets no grant.
//
// The holder defaults to each card's OWNER, per card
// (GrantCastPermissionOverCardForEffect fills a zero holder in), so
// one airbend of two players' permanents hands each player their own
// card back.
func exileAllWithPermission(g *game.Game, ids []uuid.UUID, perm game.CastPermission) error {
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, landed []uuid.UUID) error {
		for _, id := range landed {
			p := perm
			p.Zone = game.ZoneExile
			g.GrantCastPermissionOverCardForEffect(id, p)
		}
		return nil
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
	return AirbendAll{Targets: []uuid.UUID{a.Target}}.Apply(ctx)
}

// AirbendAll airbends several cards at once — "airbend any number of
// other target nonland permanents you control" (Appa). The reminder
// text says "Exile THEM", one instruction, so the cards leave as one
// simultaneous exit (a leaves-the-battlefield watcher sees the whole
// group go) rather than one exile per card.
type AirbendAll struct {
	Targets []uuid.UUID
}

func (a AirbendAll) Apply(ctx *Context) error {
	targets := ctx.withoutNewSourceObject(a.Targets) // #1432
	if len(targets) == 0 {
		return nil
	}
	return exileAllWithPermission(ctx.Game, targets, ExileWithPermission{
		CastOnly:     true,
		WhileExiled:  true,
		CostOverride: AirbendCost,
	}.permission())
}

// AirbendOtherTarget is "airbend up to one OTHER target …" — the
// trigger body Aang, the Last Airbender, Aang, Swift Savior and Avatar
// Yangchen all share: read the item's first target and decline it if
// there is none (CR 608.2b, "up to one" answering zero) or if it names
// the source itself ("other" — enforced here rather than in the
// target clause's predicate, because a TargetSpec is built once at
// Register, before an InstanceID exists to exclude). Otherwise airbend
// it.
func AirbendOtherTarget(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := item.Targets[0]
	if target.ID == item.SourceCardID {
		return nil
	}
	return Airbend{Target: target.ID}.Apply(NewContext(g, item))
}
