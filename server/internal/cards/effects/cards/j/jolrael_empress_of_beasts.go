package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jolrael, Empress of Beasts — Legendary Creature — Human Spellshaper
// {3}{G}{G}, 3/3 (Prophecy):
//
//	"{2}{G}, {T}, Discard two cards: All lands target player controls become
//	 3/3 creatures until end of turn. They're still lands."
//
// Natural Affinity's animation over the lands one player controls, locked as
// the ability resolves (CR 611.2c): a land that enters later is not animated.
// A target player who left, or an illegal target, animates nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7972d118-c105-4b33-8c3b-ec85da538471",
		Name:         "Jolrael, Empress of Beasts",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}, Discard two cards: All lands target player controls become 3/3 creatures until end of turn. They're still lands.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.DiscardN(2, "two cards")),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := effects.TargetedPlayer(ctx)
				if player == uuid.Nil {
					return nil
				}
				return effects.ScopedEffectFor{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() && c.Controller == player },
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.SetBasePowerMod(3),
						game.SetBaseToughnessMod(3),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Jolrael, Empress of Beasts — lands are 3/3 creatures",
				}.Apply(ctx)
			},
		}},
	})
}
