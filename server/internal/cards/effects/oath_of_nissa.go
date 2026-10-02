package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oath of Nissa — Legendary Enchantment {G}:
//
//	"When Oath of Nissa enters, look at the top three cards of your
//	 library. You may reveal a creature, land, or planeswalker card
//	 from among them and put it into your hand. Put the rest on the
//	 bottom of your library in any order.
//	 You may spend mana as though it were mana of any color to cast
//	 planeswalker spells."
//
// The entry trigger is the look-and-take dig (TakeFromLibraryToHand:
// a private look, an optional revealed take, the rest ordered onto the
// bottom by the player — ADR 0088).
//
// The second line is #1600's player static narrowed to one kind of
// payment: it widens the cost of casting a planeswalker spell and
// nothing else — not a loyalty ability, not another spell, not a ward
// payment (game/spend_any_color.go, AnyColorSpendStatic.Covers).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c4efdbab-711d-4269-9b24-b05d36f7e5c7",
		Name:         "Oath of Nissa",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Oath of Nissa — look at the top three cards of your library", oathOfNissaDig),
		},
		AnyColorSpend: YouMaySpendManaAsAnyColorToCast("planeswalker spells", "Planeswalker"),
	})
}

// oathOfNissaDig is the entry trigger: look at three, may take a
// creature, land or planeswalker card (revealed), the rest on the
// bottom in any order.
func oathOfNissaDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
		Match:    Or(Creature(), Land(), Planeswalker()),
		Max:      1,
		Optional: true,
		Reveal:   true,
		Label:    "Oath of Nissa — reveal a creature, land, or planeswalker card and put it into your hand",
		Then:     TakeRestOnBottomInAnyOrder,
	}.Apply(ctx)
}
