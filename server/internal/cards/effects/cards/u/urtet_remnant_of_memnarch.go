package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urtet, Remnant of Memnarch — Legendary Artifact Creature — Myr {3}, 2/2
// (March of the Machine: Phyrexia: All Will Be One Commander):
//
//	"Whenever you cast a Myr spell, create a 1/1 colorless Myr artifact
//	 creature token.
//	 At the beginning of combat on your turn, untap each Myr you control.
//	 {W}{U}{B}{R}{G}, {T}: Put three +1/+1 counters on each Myr you control.
//	 Activate only during your turn."
//
// The first ability triggers on the cast (CR 601.2i), not on resolution, so a
// countered Myr still leaves the token; the subtype is read off the spell. The
// untap and the counters each take the set of Myr the controller has as the
// ability resolves (CR 611.2c), through the layered view, so a changeling is a
// Myr. Urtet is a Myr itself and is in both sets. The activation is gated to
// the controller's own turn and taps Urtet as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6222fccf-fc08-4190-8d40-a56d6d1423df",
		Name:         "Urtet, Remnant of Memnarch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Subtype("Myr"),
				"Urtet, Remnant of Memnarch — create a 1/1 colorless Myr artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Myr artifact"), N: 1})),
			effects.AtBeginningOfYourCombat("Urtet, Remnant of Memnarch — untap each Myr you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B16UntapAllYouControlMatching(effects.NewContext(g, item), item.Controller,
						func(c game.Card) bool { return c.HasSubtypeWord("Myr") })
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{W}{U}{B}{R}{G}, {T}: Put three +1/+1 counters on each Myr you control. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.Plus(effects.ManaCost("{W}{U}{B}{R}{G}"), effects.TapCost()),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var ids []game.Card
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtypeWord("Myr") {
						ids = append(ids, c)
					}
				}
				for _, c := range ids {
					if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
						continue
					}
					if err := (effects.AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 3}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
