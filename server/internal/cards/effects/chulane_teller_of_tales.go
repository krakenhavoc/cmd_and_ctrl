package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chulane, Teller of Tales — Legendary Creature — Human Druid
// {2}{G}{W}{U}, 2/4 (EDHREC rank 2597):
//
//	"Vigilance
//	 Whenever you cast a creature spell, draw a card, then you may put
//	 a land card from your hand onto the battlefield.
//	 {3}, {T}: Return target creature you control to its owner's hand."
//
// The Bant value commander. Vigilance rides PrintedKeywords; the
// cast trigger is Beast Whisperer's condition
// (b12CreatureSpellCastByYou — the spell is read off the stack) and
// draws one; the activation is a CR 602 ability with a mana-and-tap
// cost over a creature the controller controls, bounced to its
// owner's hand on resolution — so a bounced Clone goes back to whom
// it belongs.
//
// DECLARED SIMPLIFICATION, weaker than printed: the land drop is not
// offered. "You may put a land card from your hand onto the
// battlefield" is a pick-from-hand at resolution, and the engine
// has no prompt for it (the Eureka Moment / Broken Bond posture —
// see broken_bond.go). The draw, which is the engine, is whole; the
// ramp is absent rather than automated wrongly.
func init() {
	Register(Spec{
		OracleID:        "ebf7ce9b-9e5e-4557-9e28-76556997f0ee",
		Name:            "Chulane, Teller of Tales",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The cast trigger only draws the card — it doesn't offer to put a land from your hand onto the battlefield."},
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12CreatureSpellCastByYou(ev, source, g)
			}, "Chulane, Teller of Tales — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Return target creature you control to its owner's hand.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BounceToHand{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
