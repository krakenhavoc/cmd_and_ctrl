package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b118_tokens.go — the token template of S82 set GAP, batch 118
// (#5948). Its slug is an on-disk identity: never renamed, never reused.

// S82GapB118GwaihirAttack is "Whenever <name> attacks, target attacking
// creature gains flying until end of turn." (Gwaihir and the Bird he makes).
func S82GapB118GwaihirAttack(name string) game.TriggeredAbility {
	return Targeting(WheneverThisAttacks(name+" — target attacking creature gains flying until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetCard {
				return nil
			}
			return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"flying"}, Label: name + " — flying"}.Apply(ctx)
		}),
		TargetCreature("target attacking creature", AttackingCreature()))
}

// S82GapB118GwaihirBirdToken is Gwaihir's Bird for CreateToken.
func S82GapB118GwaihirBirdToken() game.Card { return TokenFromCatalog(PrintedS82GapB118GwaihirBird) }

// PrintedS82GapB118GwaihirBird is the Bird as PRINTED (Tales of Middle-earth
// Commander token): a 3/3 white Bird with flying and "Whenever this creature
// attacks, target attacking creature gains flying until end of turn."
func PrintedS82GapB118GwaihirBird() TokenTemplate {
	return TokenTemplate{
		Slug: "bird-attacker-grants-flying",
		Card: game.Card{
			Name:      "Bird",
			TypeLine:  "Token Creature — Bird",
			Power:     3,
			Toughness: 3,
			Colors:    []string{"W"},
			Keywords:  []string{"flying"},
		},
		Triggered: []game.TriggeredAbility{
			S82GapB118GwaihirAttack("Bird"),
		},
		Text: "Flying\nWhenever this creature attacks, target attacking creature gains flying until end of turn.",
	}
}
