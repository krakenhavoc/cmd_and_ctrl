package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gwaihir, Greatest of the Eagles — Legendary Creature — Bird Noble
// {4}{W}, 5/5 (Tales of Middle-earth Commander):
//
//	"Flying
//	 Whenever Gwaihir attacks, target attacking creature gains flying until
//	 end of turn.
//	 At the beginning of each end step, if you gained 3 or more life this
//	 turn, create a 3/3 white Bird creature token with flying and
//	 "Whenever this token attacks, target attacking creature gains flying
//	 until end of turn.""
//
// The end-step clause is on every player's end step and has an intervening
// "if" (CR 603.4), checked as the step begins and again as the ability
// resolves, over the life gained this turn (the net total gained, not net
// of life lost). The Bird is a catalog token template (ADR 0083), so its
// attack trigger travels with the token; the attack trigger shared with
// Gwaihir targets any attacking creature, Gwaihir or the Bird included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77572e70-4b8b-46ae-8c8d-f71b7dd456e0",
		Name:            "Gwaihir, Greatest of the Eagles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			s82GapB118GwaihirAttack("Gwaihir, Greatest of the Eagles"),
			effects.InterveningIf(effects.CondLifeGained{AtLeast: 3},
				effects.AtEachStep(game.StepEnd, "Gwaihir, Greatest of the Eagles — create a 3/3 white Bird creature token with flying and an attack trigger",
					effects.Do(effects.CreateToken{Template: s82GapB118GwaihirBird(), N: 1}))),
		},
	})
}

// s82GapB118GwaihirAttack is "Whenever <name> attacks, target attacking
// creature gains flying until end of turn."
func s82GapB118GwaihirAttack(name string) game.TriggeredAbility {
	return effects.Targeting(effects.WheneverThisAttacks(name+" — target attacking creature gains flying until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetCard {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"flying"}, Label: name + " — flying"}.Apply(ctx)
		}),
		effects.TargetCreature("target attacking creature", effects.AttackingCreature()))
}

func s82GapB118GwaihirBird() game.Card { return effects.TokenFromCatalog(s82GapB118PrintedGwaihirBird) }

// s82GapB118PrintedGwaihirBird is the Bird as PRINTED (Tales of
// Middle-earth Commander token): a 3/3 white Bird with flying and the
// attack trigger.
func s82GapB118PrintedGwaihirBird() effects.TokenTemplate {
	return effects.TokenTemplate{
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
			s82GapB118GwaihirAttack("Bird"),
		},
		Text: "Whenever this creature attacks, target attacking creature gains flying until end of turn.",
	}
}
