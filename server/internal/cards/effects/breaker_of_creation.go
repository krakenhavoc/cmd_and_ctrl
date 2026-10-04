package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breaker of Creation — Creature — Eldrazi {6}{C}{C}, 8/4:
//
//	"When you cast this spell, you gain 1 life for each colorless
//	 permanent you control.
//	 Hexproof from each color
//	 Annihilator 2"
//
// The cast trigger resolves above the spell, so it happens even if the
// Breaker is countered (the 2024-06-07 ruling), and counts the
// colourless permanents you control as it resolves — the Breaker is
// still a spell then, so it does not count itself. Annihilator 2 is
// the canonical keyword token (ADR 0113 §2).
//
// Sandbox simplification, declared: "hexproof from each color" is a
// quality of hexproof the engine has no vocabulary for (ADR 0038 §6),
// so the Breaker can be targeted by coloured spells and abilities —
// weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:        "785026c5-3f26-489c-8e26-96dd3ca6bc98",
		Name:            "Breaker of Creation",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Hexproof from each color isn't implemented — opponents' colored spells and abilities can target it."},
		PrintedKeywords: []string{"annihilator 2"},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Breaker of Creation — gain 1 life for each colorless permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsColorless() {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
