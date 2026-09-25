package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hanweir Garrison — Creature — Human Soldier {2}{R}, 2/3 (EDHREC
// rank 4257):
//
//	"Whenever this creature attacks, create two 1/1 red Human creature
//	 tokens that are tapped and attacking.
//	 (Melds with Hanweir Battlements.)"
//
// Three mana that turns every attack into six power of damage, and
// leaves two bodies behind for the sacrifice deck. It is the token
// half of a meld pair, but the Garrison is played on its own merits —
// most decks that run it never see the Battlements.
//
// The tokens really do enter TAPPED AND ATTACKING the same player the
// Garrison is attacking: the template carries Tapped and
// AttackingTarget, CreateTokenForEffect copies both, and the combat
// damage step reads AttackingTarget off the battlefield. They were
// never DECLARED as attackers, so "whenever a creature attacks"
// triggers do not fire for them (CR 508.4) — including the Garrison's
// own, which is what keeps it from making tokens forever.
//
// The defending player is read off the attack EVENT, at trigger time,
// and carried on the item's Params — not recomputed live at
// resolution, because a live re-derivation could answer differently if
// the attacked planeswalker or battle changed hands in response, and
// CR 506.4's defending player is fixed at declaration. An attack whose
// defender has left the game between declaration and resolution makes
// no tokens rather than tokens attacking nobody.
//
// Declared simplification, weaker than printed (#259): MELD is not
// modelled. Hanweir Garrison and Hanweir Battlements will never become
// Hanweir, the Writhing Township, so the pair is two ordinary
// permanents. Nothing else about either card changes.
func init() {
	Register(Spec{
		OracleID:     "7cb29569-48e1-4782-9906-fad155ebfafe",
		Name:         "Hanweir Garrison",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"It never melds with Hanweir Battlements — the two stay separate permanents and Hanweir, the Writhing Township can't be made.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: ThisAttacked,
			Key:       "Hanweir Garrison — two tapped and attacking Humans",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Hanweir Garrison — two tapped and attacking Humans", nil)
				item.Params.Player = b17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil {
					return nil
				}
				tmpl := TokenCard("1/1 red Human")
				tmpl.Tapped = true
				tmpl.AttackingTarget = defender
				return CreateToken{Controller: item.Controller, Template: tmpl, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
