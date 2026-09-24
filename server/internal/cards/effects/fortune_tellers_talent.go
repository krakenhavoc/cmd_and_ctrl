package effects

import (
	"github.com/google/uuid"

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
// (ADR 0071), so the card registers, and #1314 closed the last two
// gaps: no simplification is left.
//
// LEVEL 1, "you may look at the top card of your library any time"
// (CR 401.5), is game.LibraryTopOwner on LibraryTopVisible. Present
// from the moment the Class enters and needing no gate: CR 716.2b
// starts every Class at level 1, so the line is never NOT there, the
// same reading `LevelUp`'s own doc comment gives the level-up
// abilities themselves.
//
// LEVEL 2, "as long as you've cast a spell this turn, you may play
// cards from the top of your library", is a `game.CastPermissionGate`
// carrying BOTH halves the printed line asks for and neither of which
// existed before #1314:
//
//   - `ActiveWhen: Level(2)` — CR 716.2a, the same gate the level-3
//     cost reduction below already uses one slot over.
//   - `Condition: hasCastASpellThisTurn` — the "as long as you've cast
//     a spell" clause, which is not a Designation (CR 716 does not
//     name it) and is asked as a second, independent test.
//
// `TopOfLibraryOnly: true` and no `CastOnly` (the card says PLAY, not
// CAST, and a card that only played never lets its owner strand a
// land the way an airbent one legitimately can). Visibility comes from
// the LEVEL 1 line above — the same permanent's own
// LibraryTopVisible — so the position check
// (`permissionPositionOKLocked`) is satisfied by this Class alone,
// exactly like every printed Talent-family card that ships both
// halves together.
//
// The level-3 cost reduction is unchanged: a `CostModifier` gated at
// `Level(3)`, the card that made cost modifiers a gateable slot in the
// first place. At level 1 or 2 the modifier is not gathered, so the
// enumerator and the engine price a graveyard cast identically — the
// #544 invariant, for free.
func init() {
	Register(Spec{
		OracleID:          "1b430f67-3686-4452-8594-b060f1a5a04e",
		Name:              "Fortune Teller's Talent",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{3}{U}")),
			LevelUp(3, ManaCost("{2}{U}")),
		},
		GatedCastPermissions: []game.CastPermissionGate{
			{
				Permission: game.CastPermission{
					Zone:             game.ZoneLibrary,
					TopOfLibraryOnly: true,
					Label:            "Fortune Teller's Talent — play cards from the top of your library",
				},
				ActiveWhen: Level(2),
				Condition:  hasCastASpellThisTurn,
			},
		},
		CostModifiers: []game.CostModifier{
			gatedCostModifier(Level(3), CostsLess(2,
				"Fortune Teller's Talent — spells you cast from anywhere other than your hand cost {2} less",
				YourSpell(), castFromOutsideYourHand())),
		},
	})
}

// hasCastASpellThisTurn is "as long as you've cast a spell this turn"
// (#1314) — the same per-turn tally EachPlayerMaxSpellsPerTurn reads,
// asked the other way round: a permission rather than a restriction.
func hasCastASpellThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).Total > 0
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
