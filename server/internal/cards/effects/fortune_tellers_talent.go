package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fortune Teller's Talent — Enchantment — Class, {U}:
//
//	(Gain the next level as a sorcery to add its ability.)
//	You may look at the top card of your library any time.
//	{3}{U}: Level 2
//	As long as you've cast a spell this turn, you may play cards from
//	the top of your library.
//	{2}{U}: Level 3
//	Spells you cast from anywhere other than your hand cost {2} less
//	to cast.
//
// #333 and ADR 0037's "blocked on" row: "A Class, not a Saga …
// levels are, and no level machinery exists." The levels exist now
// (ADR 0071), so the card registers — with one caveat, because the
// levels were never its only gap.
//
// What works:
//
//   - Both level-up abilities.
//   - Level 1's "you may look at the top card of your library any
//     time" — Realmwalker's `LibraryTopOwner` (CR 401.5): private to
//     the Talent's controller, not revealed to the table. It needs no
//     gate, because a Class is always at least level 1 (CR 716.2a).
//   - The level-3 cost reduction, a `CostModifier` gated at
//     `Level(3)`. It is the card that made cost modifiers a gateable
//     slot: everything about the reduction is ordinary except that it
//     does not exist below level 3, and the gate is the only place
//     that is said. At level 1 or 2 the modifier is not gathered, so
//     the enumerator and the engine price a graveyard cast
//     identically — the #544 invariant, for free.
//
// What does not: level 2's "as long as you've cast a spell this turn,
// you may play cards from the top of your library". Playing from the
// top of a library is NOT the gap any more — #765 closed it, and
// `PlayFromTopOfYourLibrary` is how Bolas's Citadel and Realmwalker
// do it. The gap is that a STANDING cast permission cannot be
// switched on and off: `game.standingCastPermissionsLocked` hands out
// every entry in `Spec.CastPermissions` for as long as the permanent
// is on the battlefield, and `game.CastPermission` carries neither an
// `ActiveWhen` designation (ADR 0071's gate, which statics, triggers,
// activated abilities and cost modifiers all have) nor a condition
// for "you've cast a spell this turn". Declared unconditionally, the
// permission would be live at level 1 and before the turn's first
// spell — a different, stronger card — so it stays a caveat rather
// than an approximation.
//
// Nothing here is a levels problem any more, which is why the ADR
// 0037 row is superseded rather than amended.
func init() {
	Register(Spec{
		OracleID:     "1b430f67-3686-4452-8594-b060f1a5a04e",
		Name:         "Fortune Teller's Talent",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Level 2's \"you may play cards from the top of your library\" is not implemented; play from your hand as usual.",
		},
		LibraryTopVisible: game.LibraryTopOwner,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{3}{U}")),
			LevelUp(3, ManaCost("{2}{U}")),
		},
		CostModifiers: []game.CostModifier{
			gatedCostModifier(Level(3), CostsLess(2,
				"Fortune Teller's Talent — spells you cast from anywhere other than your hand cost {2} less",
				YourSpell(), castFromOutsideYourHand())),
		},
	})
}

// castFromOutsideYourHand is "from anywhere other than your hand".
// The zone the spell is being cast from is on the query already —
// CR 601.2 makes it part of the announcement — so this is a read
// rather than provenance the permanent has to remember.
func castFromOutsideYourHand() CostPredicate {
	return func(q game.CostQuery) bool { return q.FromZone != game.ZoneHand }
}

// gatedCostModifier stamps a designation gate onto a modifier built
// with the ordinary CostsLess / CostsMore constructors, so a gated
// line reads as one thing (ADR 0071). The mirror of AtLevel /
// WhenSolved for the cost-modifier slot.
func gatedCostModifier(gate game.Designation, m game.CostModifier) game.CostModifier {
	m.ActiveWhen = gate
	return m
}
