package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swans of Bryn Argoll — Creature — Bird Spirit {2}{W/U}{W/U}, 4/3:
//
//	"Flying
//	 If a source would deal damage to this creature, prevent that damage. The source's controller draws cards equal to the damage prevented this way."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect runs
// once per SOURCE in a damage instance (PreventDamageASourceWouldDeal):
// blocked by two, each blocker's controller draws for their own creature's
// damage. The source's controller is read as the damage would have been
// dealt, from its last-known information once it has left (the ruling).
// Damage that can't be prevented draws nothing (CR 615.12: nothing was
// prevented this way). A wither source's damage is prevented like any
// other.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "b187aeeb-5cf5-4b73-a3ef-f39188d2ba33",
		Name:            "Swans of Bryn Argoll",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			PreventDamageASourceWouldDeal(PreventionStatic{
				To:    ToThisCreature,
				Then:  swansDrawBody,
				Label: "Swans of Bryn Argoll — prevent the damage; its source's controller draws that many",
			}),
		},
	})
}

// The additional effect: the source's controller (the params' Player)
// draws the damage prevented this way.
var swansDrawBody = game.DelayedBody("swans-of-bryn-argoll/source-controller-draws", swansDraw)

func swansDraw(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || p.Player == uuid.Nil {
		return nil
	}
	return DrawCards{Player: p.Player, N: p.Amount}.Apply(NewContext(g, item))
}
