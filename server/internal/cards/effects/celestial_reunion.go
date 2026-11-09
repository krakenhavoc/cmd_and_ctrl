package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Reunion — Sorcery {X}{G} (Lorwyn Eclipsed, #5406):
//
//	"As an additional cost to cast this spell, you may choose a creature
//	 type and behold two creatures of that type.
//	 Search your library for a creature card with mana value X or less,
//	 reveal it, put it into your hand, then shuffle. If this spell's
//	 additional cost was paid and the revealed card is the chosen type,
//	 put that card onto the battlefield instead of putting it into your
//	 hand."
//
// The behold is an optional additional cost that chooses its type
// (OptionalBeholdChosenType, #5406): the caster names the type on
// cast_spell's behold_type as they announce it (CR 601.2b) and two
// creatures of it on reveal_ids, creature cards in hand and creatures
// they control in any mix (CR 701.4a). Kindred cards and permanents do
// not pay it (the card's ruling). The type rides the stack item
// (PaidCost.BeholdType) and is read back here.
//
// The search is a creature card with mana value X or less (a card with
// {X} in its cost counts X as 0 in the library, CR 202.3e, which
// ManaValue already answers), revealed. With the cost paid, a found card
// of the chosen type goes to the battlefield instead of the hand
// (SearchLibrary.OntoBattlefieldIfSubtype), so it never passes through
// the hand. Then shuffle either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ba0a815c-1dfe-47c1-890f-496c670124f3",
		Name:          "Celestial Reunion",
		XMatters:      true,
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1},
		OptionalCosts: []game.AdditionalCost{OptionalBeholdChosenType(2)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			s := SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Celestial Reunion — a creature card with mana value X or less",
			}
			if t, ok := ctx.BeheldType(); ok {
				s.OntoBattlefieldIfSubtype = t
			}
			return s.Apply(ctx)
		},
	})
}
