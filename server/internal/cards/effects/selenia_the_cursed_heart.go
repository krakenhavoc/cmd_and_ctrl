package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selenia, the Cursed Heart — {4}{W}{B} Legendary Creature — Angel Warrior
// 2/5 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Flying, lifelink
//	 If you would gain life, you gain twice that much life instead.
//	 When Selenia dies, create a legendary black Aura Curse enchantment
//	 token named Selenia's Curse attached to target opponent. The token
//	 has enchant player and "If enchanted player would lose life, they
//	 lose twice that much life instead.""
//
// The doubler is Rhox Faithmender's (YouGainTwiceThatMuchLife), and
// reaches the printed lifelink. The dies trigger targets an opponent as
// it goes on the stack and is re-checked as it resolves; if the opponent
// is gone nothing is created. The token is created already attached to
// the player (TokenEntryOptions.AttachedTo, CR 303.4i) and controlled by
// Selenia's controller; its replacement is a token-template replacement
// (tokenTemplate.Replacements) that doubles any life loss by the player
// it is attached to, damage included (CR 120.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7126be80-cebc-4868-b39d-8b215410f575",
		Name:            "Selenia, the Cursed Heart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Replacements: []game.ReplacementEffect{
			YouGainTwiceThatMuchLife("Selenia, the Cursed Heart: gain twice that much life"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Selenia, the Cursed Heart — create Selenia's Curse attached to target opponent",
					seleniaCurse),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func seleniaCurse(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(SeleniasCurseToken()), 1, game.TokenEntryOptions{
			AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: t.ID},
		})
		return err
	}
	return nil
}
