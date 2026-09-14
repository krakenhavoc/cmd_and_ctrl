package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Griselbrand — Legendary Creature — Demon 7/7 for {4}{B}{B}{B}{B}:
//
//	"Flying, lifelink
//	 Pay 7 life: Draw seven cards."
//
// The reanimation target, and the card that makes the life total a
// library. Both keywords ride PrintedKeywords; the ability is a
// pure life cost with no mana and no tap component, so it can be
// activated any number of times in a row and at instant speed —
// that is the card, not an oversight.
//
// Lifelink and the ability compose the way the rules say and the
// engine already handles: the seven life is paid at announce
// (CR 601.2h), so a Griselbrand at 7 life may activate once and is
// then at 0, losing to the state-based check — and a Griselbrand
// that connected for 7 first is back where it started, which is why
// the pair is the format's defining combo piece rather than two
// unrelated abilities.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "f759d112-76db-4091-a22b-b9f19ab6fa5f",
		Name:            "Griselbrand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Activated: []ActivatedAbility{{
			Label: "Pay 7 life: Draw seven cards",
			Cost:  PayLife(7),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 7}.Apply(NewContext(g, item))
			},
		}},
	})
}
