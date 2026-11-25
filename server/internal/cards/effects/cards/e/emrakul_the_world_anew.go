package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emrakul, the World Anew — Legendary Creature — Eldrazi {12}, 12/12
// (Modern Horizons 3, S82 GAP2 b11, #6098):
//
//	"When you cast this spell, gain control of all creatures target player
//	 controls.
//	 Flying, protection from spells and from permanents that were cast
//	 this turn
//	 When Emrakul leaves the battlefield, sacrifice all creatures you
//	 control.
//	 Madness—Pay six {C}."
//
// The cast trigger fires from the stack (CR 601.2i, 603.2) and resolves
// above the spell, so it resolves even if Emrakul is countered. It takes
// every creature the target player controls as it resolves, for good
// (Karrthus's indefinite control change, CR 611.2c locks the set then).
//
// The protection is two protection abilities (CR 702.16g): "spells" is any spell (an
// ability is not), and "permanents that were cast this turn" is a
// permanent whose spell was cast this turn (CastProvenance.CastThisTurn),
// so a creature that was put onto the battlefield, or was cast an earlier
// turn, can block it and target it. Both answer targeting, attaching,
// damage and blocking (CR 702.16b-f).
//
// The leaves trigger sacrifices every creature its controller controls as
// it resolves, as one simultaneous event (B08SacrificeAllMatching).
//
// Madness (CR 702.35) with a cost of six colorless mana, which only {C}
// pays (CR 107.4c).
//
// No simplification.
func init() {
	castTrigger := effects.WhenYouCastThisSpell("Emrakul, the World Anew — gain control of all creatures target player controls", emrakulTheWorldAnewSteal)
	castTrigger.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "33300e69-29f6-41e4-92ab-82d8b018b217",
		Name:         "Emrakul, the World Anew",
		Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{
			"flying",
			game.KeywordProtection + " from spells",
			game.KeywordProtection + " from permanents that were cast this turn",
		},
		Madness: "{C}{C}{C}{C}{C}{C}",
		Triggered: []game.TriggeredAbility{
			castTrigger,
			effects.WhenThisLeaves("Emrakul, the World Anew — sacrifice all creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B08SacrificeAllMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.YouControl()))
				}),
		},
	})
}

func emrakulTheWorldAnewSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
			if c.Controller != t.ID {
				continue
			}
			if err := (effects.GainControl{
				Target:   c.InstanceID,
				Duration: game.IndefiniteDuration(),
				Label:    "Emrakul, the World Anew — gain control of all creatures target player controls",
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
