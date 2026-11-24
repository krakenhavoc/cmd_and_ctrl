package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Howling Abomination — Legendary Creature — Human Beast Warrior
// {3}{R}{G}, 5/5 (Secret Lair Drop, #5462):
//
//	"Haste
//	 The Howling Abomination has trample as long as you've cast three or more
//	 spells this turn.
//	 Whenever The Howling Abomination becomes the target of a spell, it gets
//	 +2/+2 until end of turn and deals 2 damage to each opponent."
//
// The trample is a layer 6 grant on itself gated on its controller's cast
// tally for the turn (Brightspear Zealot's gate, DependsOnSpellsCast), so it
// switches on the moment the third spell is cast. The trigger watches targets
// of SPELLS only (SelfTargetedByASpell, Bonecrusher Giant's): an ability that
// targets it does nothing. It fires at announce (CR 601.2c), goes on the
// stack above the spell, and does both things as it resolves, the damage
// dealt by the Abomination as it last existed.
//
// No simplification.
func init() {
	trample := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && g.CastTallyFor(source.Controller).Total >= 3
	}, "trample")
	trample.Label = "The Howling Abomination has trample as long as you've cast three or more spells this turn."
	trample.DependsOnSpellsCast = true
	effects.Register(effects.Spec{
		OracleID:        "4a67946e-8903-4994-9641-d31fbb62082c",
		Name:            "The Howling Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{trample},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.SelfTargetedByASpell,
				"The Howling Abomination — it gets +2/+2 until end of turn and deals 2 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(2, 2, "The Howling Abomination — +2/+2 until end of turn")(g, item); err != nil {
						return err
					}
					return effects.DamageToEachOpponent(g, item, 2)
				}),
		},
	})
}
