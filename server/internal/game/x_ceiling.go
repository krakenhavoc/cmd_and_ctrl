package game

import "github.com/google/uuid"

// x_ceiling.go — a printed ceiling on a spell's announced X, set by a
// count on the board (#2581).
//
//	Winter's Chill  X can't be greater than the number of snow lands
//	                you control.
//	Open the Way    X can't be greater than the number of players in
//	                the game.
//
// CR 107.3a makes X a number the caster chooses and announces, and CR
// 601.2b says when: as part of announcing the spell, before targets
// (601.2c) and before the total cost is determined and paid
// (601.2f–h). A printed "X can't be greater than N" is a rule of the
// card that bounds that announcement, so N is read once, at that
// moment. Nothing re-reads it: the spell's X is the announced value
// for as long as it is on the stack (CR 107.3a), so a snow land that
// leaves in response, or a player who leaves the game (Open the Way's
// ruling of 2023-05-12), changes nothing about the X already chosen.
//
// The two existing X bounds are costs, and this is not one. "Pay X
// life" (AdditionalCost.PayLifeX) and "blight X"
// (AdditionalCost.BlightX) cap X because X is what they CHARGE; the
// ceiling here charges nothing and only limits the choice. It joins
// them in the same places, so all of them are read alike:
//
//   - the announce gate in castSpellLocked refuses an X above it;
//   - the legal-move enumerator folds it into the one X ceiling it
//     already solves against (internal/legal/x.go), so neither the bot
//     nor the MCP seat's open X range ever exceeds it;
//   - the view stamps it on the caster's own card as `x_max`, and the
//     auto-tap preview reports it, so the client's X picker stops at it.
//
// One function, SpellXCeilingLocked, answers all four, so they cannot
// disagree about the number.
//
// Spells only. No activated ability in the catalog prints a board-set
// ceiling; an activation's printed floor is AbilityCost.MinX.

// XCeiling is a printed "X can't be greater than <Label>" on a spell.
//
// Catalog data on the card's definition (CardDef), never on game state:
// the count is read once at announce and the announced X is what the
// stack item records, so nothing about it needs to survive a restore.
type XCeiling struct {
	// Label is the count's printed words — "the number of snow lands
	// you control" — for the refusal log and the client's hint.
	Label string
	// Count is the ceiling right now for a spell `caster` is casting.
	// Read-only, under g.mu with the layers fresh; it must not call a
	// public locking accessor. A negative answer is read as 0
	// (CR 107.1b).
	Count func(g *Game, caster uuid.UUID) int
}

// CatalogXCeiling is the catalog hook for a card's printed X ceiling;
// carddef.go's init reads it off the CardDef. Nil (no catalog) means no
// card declares one.
var CatalogXCeiling func(oracleID string) *XCeiling

// XCeilingFor is the printed X ceiling the catalog declares for `key`,
// or nil for the overwhelming majority of cards, which print none.
func XCeilingFor(key string) *XCeiling {
	if CatalogXCeiling == nil || key == "" {
		return nil
	}
	return CatalogXCeiling(key)
}

// SpellXCeilingLocked is the largest X `caster` may announce for the
// spell whose catalog key is `key`, read now (CR 601.2b). ok is false
// when the card prints no ceiling — X is then bounded by its costs
// alone.
//
// The count reads effective characteristics (a land an effect made
// snow is snow), so the layers must be fresh: castSpellLocked
// recomputes them first, ReadSnapshot and the view run fresh, and
// SpellXCeiling below goes through ReadSnapshot.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) SpellXCeilingLocked(caster uuid.UUID, key string) (ceiling int, ok bool) {
	xc := XCeilingFor(key)
	if xc == nil || xc.Count == nil {
		return 0, false
	}
	return max(0, xc.Count(g, caster)), true
}

// SpellXCeiling is SpellXCeilingLocked for a caller holding no lock —
// the auto-tap preview handler, which reports the ceiling beside the
// price so the client's X picker stops where CastSpell does.
func (g *Game) SpellXCeiling(caster uuid.UUID, key string) (ceiling int, ok bool) {
	g.ReadSnapshot(func() { ceiling, ok = g.SpellXCeilingLocked(caster, key) })
	return ceiling, ok
}

// SpellXCeilingForEffect is SpellXCeilingLocked on the *ForEffect
// surface, for the legal-move enumerator and the protocol view, which
// must quote the ceiling the announce gate enforces.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) SpellXCeilingForEffect(caster uuid.UUID, key string) (int, bool) {
	return g.SpellXCeilingLocked(caster, key)
}
