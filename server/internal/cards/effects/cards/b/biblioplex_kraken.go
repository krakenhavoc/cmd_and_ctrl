package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biblioplex Kraken — Creature — Kraken {4}{U}, 4/4 (Jumpstart 2022, #5948):
//
//	"When this creature enters, scry 3.
//	 Whenever this creature attacks, you may return another creature you
//	 control to its owner's hand. If you do, this creature can't be blocked
//	 this turn."
//
// Temur Sabertooth's "you may return another creature you control; if you do"
// (nothing targets, zero picks is the decline, the source is never offered) on
// an attack trigger. The unblockable rider runs behind the bounce, only if the
// creature really reached a hand; the attacking Kraken is already declared, so
// it is judged at the declare-blockers step.
//
// No simplification.
func init() {
	const label = "Biblioplex Kraken — you may return another creature you control to its owner's hand; if you do, it can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "a1bdba37-f034-412f-8395-adbaed95ad11",
		Name:         "Biblioplex Kraken",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Biblioplex Kraken — scry 3", effects.Do(effects.Scry{N: 3})),
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				return effects.ChooseAnotherCreatureYouControl(g, item,
					"Biblioplex Kraken — return another creature you control to its owner's hand?", biblioplexKrakenReturn)
			}),
		},
	})
}

// biblioplexKrakenReturn bounces the creature the Kraken's controller picked
// and, behind the move, makes the Kraken unblockable this turn. On-disk keys.
var biblioplexKrakenReturn = effects.RegisterPermanentsPick("prompt-run/biblioplex-kraken-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			return effects.BounceToHand{Target: id, ThenTail: ctx.ItemTail(biblioplexKrakenUnblockable)}.Apply(ctx)
		}
		return nil
	},
})

var biblioplexKrakenUnblockable = effects.CardTail("biblioplex-kraken/unblockable", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || !effects.OnBattlefield(ctx.Game, ctx.Source()) || ctx.Game.AbilitySourceIsNewObjectForEffect(ctx.Item) {
		return nil
	}
	return effects.RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked, Label: "Biblioplex Kraken — can't be blocked this turn"}.Apply(ctx)
})
