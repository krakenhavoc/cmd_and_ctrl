package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evie Frye — Legendary Creature — Human Assassin {1}{U}, 2/1 (Assassin's
// Creed, #5948):
//
//	"Partner with Jacob Frye (When this creature enters, target player may put
//	 Jacob into their hand from their library, then shuffle.)
//	 {1}, {T}: Draw a card, then discard a card. When you discard a creature
//	 card this way, target creature you control can't be blocked this turn."
//
// Partner with is the shared entry trigger (partner_with.go). The loot is
// Glacial Dragonhunt's draw-then-discard (the discard is not optional), and
// the reflexive trigger (CR 603.12) is raised only when the card discarded
// really was a creature card, its target chosen as it goes on the stack and
// given "can't be blocked" until end of turn (RestrictUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "248cafe8-cc37-4cf3-8239-ac10e55ced13",
		Name:         "Evie Frye",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Evie Frye", "Jacob Frye"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card. When you discard a creature card this way, target creature you control can't be blocked this turn.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
					Ref: evieFryeDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
				})
			},
		}},
	})
}

var (
	// The loot's discard half, run once the draw has landed. An on-disk key.
	evieFryeDiscard = game.RegisterDrawThen("evie-frye/discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			Question: "Evie Frye — discard a card",
		}, evieFryeDiscarded, nil)
	})

	// The discard's answer: a creature card discarded raises the reflexive
	// trigger. An on-disk key.
	evieFryeDiscarded = game.RegisterDiscardThen("choose-cards/evie-frye-discarded", func(g *game.Game, r game.DiscardsDone) error {
		creature := false
		for _, id := range r.Discarded.Cards() {
			if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
				creature = true
			}
		}
		if !creature {
			return nil
		}
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Evie Frye"}
		return effects.WhenYouDo("Evie Frye — target creature you control can't be blocked this turn", evieFryeUnblockableBody).Apply(effects.NewContext(g, parent))
	})

	// "When you discard a creature card this way, target creature you control
	// can't be blocked this turn." An on-disk key.
	evieFryeUnblockableBody = game.ReflexiveBody("evie-frye/unblockable", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Evie Frye — can't be blocked this turn"}.Apply(ctx)
	}), effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))
)
