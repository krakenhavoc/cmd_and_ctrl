package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arboreal Alliance — Enchantment {X}{G}{G}:
//
//	"When this enchantment enters, create an X/X green Treefolk
//	 creature token.
//	 Whenever you attack with one or more Elves, populate. (Create a
//	 token that's a copy of a creature token you control.)"
//
// The enters trigger reads the announced X off the permanent's record
// of the spell that became it (CR 107.3m, 400.7d) as the trigger is
// built, like The Goose Mother's, and builds the token by hand because
// the token table has no row for a size decided at resolution. X = 0
// makes no token (a 0/0 would die at once).
//
// The attack trigger is the "one or more" shape: the engine emits one
// attack event per attacker and OncePerBatch keeps the first of a
// combat's declaration, as Hermes' does. Subtypes are read effective,
// so a changeling attacker is an Elf.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f497352-c812-48c4-a922-cc311bcdcdc6",
		Name:         "Arboreal Alliance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Arboreal Alliance — create an X/X green Treefolk",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Arboreal Alliance — create an X/X green Treefolk")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template: game.Card{
							Name:      "Treefolk",
							TypeLine:  "Token Creature — Treefolk",
							Colors:    []string{"G"},
							Power:     x,
							Toughness: x,
						},
						N: 1,
					}.Apply(NewContext(g, item))
				},
			},
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasSubtype("Elf")
			}, "Arboreal Alliance — populate", Do(Populate{}))),
		},
	})
}
