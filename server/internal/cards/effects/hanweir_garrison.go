package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// The reminder line is the other half of Hanweir Battlements' meld
// ability (CR 701.42, 712.5b; ADR 0145, #2699): the Battlements' {3}{R}{R}
// exiles both and returns them as Hanweir, the Writhing Township. The
// attack trigger is shared with the Township
// (whenThisAttacksTokensTappedAndAttacking, meld.go).
func init() {
	Register(Spec{
		OracleID:     "7cb29569-48e1-4782-9906-fad155ebfafe",
		Name:         "Hanweir Garrison",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenThisAttacksTokensTappedAndAttacking("Hanweir Garrison — two tapped and attacking Humans", "1/1 red Human", 2),
		},
	})
}
