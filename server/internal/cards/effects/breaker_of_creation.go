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
// "Hexproof from each color" is game.HexproofFromEachColor (#3172,
// ADR 0038's amendment of 2026-10-13): an opponent's coloured spell or
// ability can't target it, a colourless one (an artifact's ability, an
// Eldrazi) still can.
func init() {
	Register(Spec{
		OracleID:        "785026c5-3f26-489c-8e26-96dd3ca6bc98",
		Name:            "Breaker of Creation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 2", game.HexproofFromEachColor},
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
