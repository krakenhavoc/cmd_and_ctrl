package game

import "github.com/google/uuid"

// CondCastASpellThisTurn is "as long as you've cast another spell this turn"
// (Monk Class's level-3 exiled card, "You may cast this card from exile as long
// as you've cast another spell this turn"). The card with the permission is in
// exile, so it is not yet one of the spells cast: any spell its holder has cast
// this turn is another one (CR 601.2a, 601.2i: a spell is cast, and counted,
// once it is on the stack with the costs paid). A pure read of the turn's cast
// tally. An on-disk key: never renamed, never reused.
const CondCastASpellThisTurn PermissionConditionKey = "cast_a_spell_this_turn"

func init() {
	RegisterPermissionCondition(CondCastASpellThisTurn, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, _ *PermissionCondition) bool {
			return g.CastTallyFor(holder).Total > 0
		},
		Label: func(*PermissionCondition) string { return "as long as you've cast another spell this turn" },
	})
}
