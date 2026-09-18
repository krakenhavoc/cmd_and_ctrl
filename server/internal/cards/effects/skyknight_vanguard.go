package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyknight Vanguard — Creature — Human Knight {R}{W}, 2/2 (EDHREC
// rank 3861):
//
//	"Flying
//	 Whenever this creature attacks, create a 1/1 white Soldier
//	 creature token that's tapped and attacking."
//
// Adeline's second ability at one seat's worth of scale: where
// Adeline reads "whenever YOU attack" and makes one token per
// opponent, this reads "whenever THIS CREATURE attacks" and makes
// exactly one, for the player this creature is already attacking.
// The pair is worth having in the catalog together — they are the two
// halves of the same template and they take different triggers,
// WheneverThisAttacks against the OncePerBatch dedup, which is how
// you find out the distinction is real.
//
// The token enters TAPPED AND ATTACKING the same defender this
// creature is attacking: the template carries Tapped and
// AttackingTarget, CreateTokenForEffect copies both, and the damage
// step reads AttackingTarget off the battlefield. It was never
// DECLARED as an attacker, so "whenever a creature attacks" triggers
// do not fire for it (CR 508.4), as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9180f77f-c288-4a24-a35d-a16270b3d737",
		Name:            "Skyknight Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Skyknight Vanguard — a tapped and attacking Soldier",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					tmpl := TokenCard("1/1 white Soldier")
					tmpl.Tapped = true
					tmpl.AttackingTarget = skyknightVanguardDefender(g, ctx.Source())
					return CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(ctx)
				}),
		},
	})
}

// skyknightVanguardDefender is the player this Vanguard is attacking,
// which is the one its token joins. Read from the source's own
// AttackingTarget rather than from the event, so the token lands on
// the right seat at a four-player table.
//
// uuid.Nil when the Vanguard is no longer on the battlefield to ask —
// killed in response to its own trigger — and a token with no
// AttackingTarget enters tapped but not attacking. That is the weaker
// reading of a card whose attacker is gone, which is the direction
// #259 asks for.
func skyknightVanguardDefender(g *game.Game, source uuid.UUID) uuid.UUID {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return uuid.Nil
	}
	return c.AttackingTarget
}
