package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const copycrookConniveGrant = "copycrook/attacks-connives"

// Copycrook — Creature — Shapeshifter Rogue {2}{U}{U}, 0/0 (Modern
// Horizons 3):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has "Whenever this creature attacks, it
//	 connives." (Draw a card, then discard a card. If you discarded a
//	 nonland card, put a +1/+1 counter on this creature.)"
//
// Colossal Clone Master's shape: EntersAsCopyOf over any creature with the
// exception a copy grant. The granted trigger is Kang's own "whenever this
// creature attacks, it connives" (WheneverThisAttacks + sourceConnives),
// part of the copy's copiable values (CR 707.9a), so a later copy of the
// Copycrook has it too. Declining leaves the printed 0/0, which dies to CR
// 704.5f.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce184565-71fb-4e23-b45b-8786727d21f7",
		Name:         "Copycrook",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       copycrookConniveGrant,
			Text:      `Whenever this creature attacks, it connives.`,
			Triggered: []game.TriggeredAbility{effects.WheneverThisAttacks("Copycrook — it connives", effects.Do(effects.Instruction(effects.SourceConnives)))},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Copycrook",
				effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(copycrookConniveGrant)
				},
			),
		},
	})
}
