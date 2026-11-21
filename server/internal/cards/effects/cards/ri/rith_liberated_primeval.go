package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rith, Liberated Primeval — Legendary Creature — Dragon {2}{R}{G}{W},
// 5/5 (Dominaria United, #5663):
//
//	"Flying, ward {2}
//	 Other Dragons you control have ward {2}.
//	 At the beginning of your end step, if a creature or planeswalker an
//	 opponent controlled was dealt excess damage this turn, create a 4/4
//	 red Dragon creature token with flying."
//
// Ward {2} is the printed trigger (Ward), and the grant to the other
// Dragons is WardGranted over the Dragon subtype (a word of its text,
// CR 612.2). The printed line says "Dragons", not "Dragon creatures",
// so a Dragon permanent that is not a creature is covered too.
//
// The end-step trigger is an intervening "if" (CR 603.4) over the turn's
// excess-damage tally (excess_damage_this_turn.go): checked as the end
// step begins and again as the trigger resolves. Excess damage is damage
// past lethal for a creature and past loyalty for a planeswalker
// (CR 120.10), combat or not, from any source; Rith need not have dealt
// it nor been on the battlefield when it was dealt, and the permanent
// need not still be around. One token however many permanents were dealt
// excess damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f89de40f-52d3-44b1-8faa-6ddaa32880d0",
		Name:            "Rith, Liberated Primeval",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "ward"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Rith, Liberated Primeval — ward {2}"),
			effects.WardGranted(effects.WardMana("{2}"), "Rith, Liberated Primeval — other Dragons you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
						target.HasSubtype(source.TextSubtype("Dragon"))
				}),
			effects.AtYourEndStepIfOpponentPermanentDealtExcessDamage(
				"Rith, Liberated Primeval — create a 4/4 red Dragon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1})),
		},
	})
}
