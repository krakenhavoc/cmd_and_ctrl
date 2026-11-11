package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clavileño, First of the Blessed — Legendary Creature — Vampire Cleric
// {1}{W}{B}, 2/2 (Lost Caverns of Ixalan Commander, #5368):
//
//	"Whenever you attack, target attacking Vampire that isn't a Demon
//	 becomes a Demon in addition to its other types. It gains "When this
//	 creature dies, draw a card and create a tapped 4/3 white and black
//	 Vampire Demon creature token with flying.""
//
// "Whenever you attack" is once per declaration, whatever the number of
// attackers (OncePerBatch). The target is chosen as the trigger goes on
// the stack (CR 603.3d): an attacking Vampire (read through the layers)
// that isn't already a Demon; with none the trigger is removed. The type
// and the ability are one effect with no stated duration, so they last
// until the permanent leaves (CR 611.2a), pinned to the object, and the
// type is added to the others (CR 205.1b). The granted dies trigger is
// the creature's own, so it draws and makes the tapped Demon only when
// that creature dies, from last-known information.
//
// No simplification.
const clavilenoDemonDies = "clavileno-first-of-the-blessed/demon-dies"

func init() {
	Register(Spec{
		OracleID:     "2b62a716-7831-4c76-8f2d-b99b04e33350",
		Name:         "Clavileño, First of the Blessed",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: clavilenoDemonDies,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Clavileño — draw a card and create a tapped 4/3 white and black Vampire Demon creature token with flying",
					Do(DrawCards{N: 1},
						CreateTokenAdvanced{Spec: Token(TokenCard("4/3 white and black Vampire Demon with flying")).EntersTapped(), N: 1})),
			},
			Text: "When this creature dies, draw a card and create a tapped 4/3 white and black Vampire Demon creature token with flying.",
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Clavileño — target attacking Vampire that isn't a Demon becomes a Demon and gains a dies trigger",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						target, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantAbilitiesFor{
							Target:   target,
							Keys:     []string{clavilenoDemonDies},
							Also:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Demon"))},
							Duration: g.PinnedTo(game.IndefiniteDuration(), target),
							Label:    "Clavileño — it becomes a Demon and gains the dies trigger",
						}.Apply(ctx)
					}),
				TargetCreature("target attacking Vampire that isn't a Demon",
					AttackingCreature(), OfCreatureType("Vampire"), Not(OfCreatureType("Demon"))))),
		},
	})
}
