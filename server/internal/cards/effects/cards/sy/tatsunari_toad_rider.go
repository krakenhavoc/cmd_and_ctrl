package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tatsunari, Toad Rider — {2}{B} Legendary Creature — Human Ninja 3/3
// (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Whenever you cast an enchantment spell, if you don't control a creature
//	 named Keimi, create Keimi, a legendary 3/3 black and green Frog
//	 creature token with "Whenever you cast an enchantment spell, each
//	 opponent loses 1 life and you gain 1 life."
//	 {1}{G/U}: Tatsunari and target Frog you control can't be blocked this
//	 turn except by creatures with flying or reach."
//
// The trigger has an intervening "if" (CR 603.4), checked as the spell is
// cast and again as the ability resolves, by name through the layers. Keimi
// is a catalog token (s82_neo_b20_tokens.go) that carries its own cast
// trigger. The activated ability's target Frog is chosen on activation; both
// creatures are pinned when it resolves (CR 611.2c), and with the Frog gone
// Tatsunari still gets the rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e87be082-eafd-480e-b150-d27fd937e1b1",
		Name:         "Tatsunari, Toad Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(tatsunariNoKeimi,
				effects.On(game.EventCast, effects.AllOf(effects.YouCast(effects.Enchantment()), func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.YouControlNoCreatureNamed(g, source.Controller, "Keimi")
				}),
					"Tatsunari, Toad Rider — if you don't control a creature named Keimi, create Keimi, a legendary 3/3 black and green Frog",
					effects.Do(effects.CreateToken{Template: effects.KeimiToken(), N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{G/U}: Tatsunari and target Frog you control can't be blocked this turn except by creatures with flying or reach.",
			Purpose: func() game.Purpose {
				p := effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))
				p.Answers = game.AnswerCombatGrant
				return p
			}(),
			Cost:    effects.ManaCost("{1}{G/U}"),
			Targets: effects.TargetCreature("target Frog you control", effects.HasSubtype("Frog"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ids := []uuid.UUID{item.SourceCardID}
				if f, ok := effects.B16FirstLegalTargetCard(ctx); ok && f != item.SourceCardID {
					ids = append(ids, f)
				}
				for _, id := range ids {
					if !effects.OnBattlefield(g, id) {
						continue
					}
					if err := (effects.CantBeBlockedThisTurnExceptBy{
						Target:   id,
						Keywords: []string{"flying", "reach"},
						Text:     "creatures with flying or reach",
						Label:    "Tatsunari, Toad Rider — can't be blocked except by creatures with flying or reach",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

var tatsunariNoKeimi = effects.CondBoard{
	Name: "you don't control a creature named Keimi",
	Test: func(ctx *effects.Context) bool {
		return effects.YouControlNoCreatureNamed(ctx.Game, ctx.Controller(), "Keimi")
	},
}
