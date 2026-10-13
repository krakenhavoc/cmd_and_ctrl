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
// The controller chooses what each Cat attacks (CR 508.4) through
// CreateTokensAttackingYourChoice, asking about what the Warleader
// attacked first (read off the attack event). No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b1351e6-165e-4ca3-96d5-4774b3176362",
		Name:         "Leonin Warleader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Key: "Leonin Warleader — two 1/1 lifelink Cats, tapped and attacking",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateTokensAttackingYourChoice{
					Template: b31TappedCatLifelinkToken(),
					N:        2,
					Tapped:   true,
					Prefer:   triggerAttackTarget(item),
					Label:    "Leonin Warleader",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
