package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saavik, Stoic Student — Legendary Creature — Vulcan Advisor {4}{W},
// 2/2 (Star Trek, #4973):
//
//	"Flash
//	 Federation — When Saavik enters, she deals X damage to target
//	 attacking or blocking creature an opponent controls, where X is the
//	 number of creature types among non-Borg creatures you control."
//
// Federation is an ability word (no rules meaning). X is counted as the
// trigger resolves (CR 608.2h), over every creature type a non-Borg
// creature you control has after the layers (trekB4CreatureTypesAmongNonBorg);
// Saavik counts herself while she is still there. The damage source is
// Saavik, read from her last-known information if she has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53aaa919-d08f-4b76-8583-bc2d4635f991",
		Name:            "Saavik, Stoic Student",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Saavik — Federation: X damage to target attacking or blocking creature an opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 {
							return nil
						}
						x := trekB4CreatureTypesAmongNonBorg(g, item.Controller, g.TextReaderForEffect(item).TextSubtype("Borg"))
						return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: x}.Apply(ctx)
					}),
				TargetCreature("target attacking or blocking creature an opponent controls",
					AttackingOrBlocking(), OpponentControls())),
		},
	})
}
