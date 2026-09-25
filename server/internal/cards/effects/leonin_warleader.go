package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Warleader — Creature — Cat Soldier {2}{W}{W}, 4/4 (EDHREC
// rank 3280):
//
//	"Whenever this creature attacks, create two 1/1 white Cat
//	 creature tokens with lifelink that are tapped and attacking."
//
// Eight power and two lifelinkers for four, every combat. The
// trigger is attackDeclared — the Warleader itself, once per attack
// — and the Cats enter TAPPED AND ATTACKING through
// CreateTokensAttackingForEffect (General Kreat's path): put onto
// the battlefield attacking, never declared, so they fire no
// "whenever a creature attacks" trigger of their own (CR 508.4) and
// can be blocked and deal their damage this combat. The Cat is Regal
// Caracal's lifelink Cat with the tapped stamp.
//
// The defending player is read off the attack EVENT, at trigger time,
// and carried on the item's Params — not recomputed live at
// resolution, because a live re-derivation could answer differently
// if the attacked planeswalker or battle changed hands in response,
// and CR 506.4's defending player is fixed at declaration.
//
// Sandbox simplification, declared: the printed card lets the
// controller choose which player (or planeswalker) each token
// attacks; here both attack the player the Warleader was declared
// against — the player behind a planeswalker or battle it attacked.
// Weaker than printed (the choice is absent), never stronger.
func init() {
	Register(Spec{
		OracleID:     "8b1351e6-165e-4ca3-96d5-4774b3176362",
		Name:         "Leonin Warleader",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Cat tokens attack the player Leonin Warleader attacked rather than a player of your choice."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Key: "Leonin Warleader — two 1/1 lifelink Cats, tapped and attacking",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Leonin Warleader — two 1/1 lifelink Cats, tapped and attacking", nil)
				item.Params.Player = b17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.CreateTokensAttackingForEffect(item.Controller, b31TappedCatLifelinkToken(), 2, item.Params.Player)
			},
		}},
	})
}
