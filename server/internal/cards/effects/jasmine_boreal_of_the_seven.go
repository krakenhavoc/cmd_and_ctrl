package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jasmine Boreal of the Seven — Legendary Creature — Human Druid
// {1}{G}{W}, 2/4:
//
//	"{T}: Add {G}{W}. Spend this mana only to cast creature spells with no
//	 abilities.
//	 Creatures you control with no abilities can't be blocked by creatures
//	 with abilities."
//
// The mana is two fixed colours carrying three AND-ed restrictions (CR
// 106.6): a cast, of a creature spell, with no abilities
// (game.ManaRestrictSpellNoAbilities reads the card's printed text, or a
// face-down cast). An activation or any other spell can't be paid with it.
//
// The block rule reads both creatures live, as the block is declared (CR
// 509.1b): the attacker is a creature Jasmine's controller controls with
// no abilities (Card.HasNoAbilities: no keyword, no granted ability, and
// no printed ability an ability-removing effect hasn't taken away), the
// blocker is any creature that has an ability (CR 113.1). A creature that
// gains an ability, even a keyword, can be blocked by anything; one that
// loses them all can't be blocked by creatures with abilities. Jasmine's
// own text is an ability, so she herself is not covered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "138201c0-556d-4c76-b383-90fc9dad0fe0",
		Name:         "Jasmine Boreal of the Seven",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{W}",
			Label:    "Add {G}{W}. Spend this mana only to cast creature spells with no abilities",
			Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictType("Creature"),
				game.ManaRestrictSpellNoAbilities},
		}},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnMatching(And(Creature(), YouControl(), creatureWithNoAbilities())),
				And(Creature(), Not(creatureWithNoAbilities())), "creatures with abilities"),
		},
	})
}

// creatureWithNoAbilities is "with no abilities" for a permanent as it is
// now (Jasmine Boreal of the Seven).
func creatureWithNoAbilities() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasNoAbilities() }
}
