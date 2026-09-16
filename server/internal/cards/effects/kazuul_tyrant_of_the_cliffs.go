package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kazuul, Tyrant of the Cliffs — Legendary Creature — Ogre Warrior
// {3}{R}{R}, 5/4 (EDHREC rank 1870):
//
//	"Whenever a creature an opponent controls attacks, if you're the
//	 defending player, create a 3/3 red Ogre creature token unless
//	 that creature's controller pays {3}."
//
// The pillow-fort Ogre: every creature that swings at you costs its
// controller {3} or hands you a 3/3. One trigger per attacking
// creature, as printed — the engine emits EventAttack per creature,
// which is exactly this card's wording — gated on the source's
// controller being the defending player (S27: an attack at your
// planeswalker or your battle is an attack you defend). The "unless"
// is the CR 118.12 PayUnless prompt addressed to the ATTACKER, and
// the Ogre is made when they decline or cannot pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8bf3d91-cb50-48ce-88c6-cdcb37b64b57",
		Name:         "Kazuul, Tyrant of the Cliffs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17OpponentsCreatureAttackedYou(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				attacker := ev.Actor
				return game.NewTriggeredItem(source, "Kazuul — a 3/3 Ogre unless the attacker pays {3}",
					func(g *game.Game, item *game.StackItem) error {
						return PayUnless{
							Chooser:  attacker,
							Cost:     "{3}",
							Question: "Kazuul, Tyrant of the Cliffs — pay {3} to stop the Ogre?",
							OnDecline: func(ctx *Context) error {
								return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/3 red Ogre"), N: 1}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
