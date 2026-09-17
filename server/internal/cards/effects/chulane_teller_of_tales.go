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
// (creatureSpellCastByYou — the spell is read off the stack) and
// draws one; the activation is a CR 602 ability with a mana-and-tap
// cost over a creature the controller controls, bounced to its
// owner's hand on resolution — so a bounced Clone goes back to whom
// it belongs.
//
// The land drop is the shared MayPutALandFromHand clause (#654),
// sequenced after the draw exactly as printed, so a land just drawn
// is a legal pick. It is a PUT, not a play (CR 305.4): Chulane ramps
// on top of the turn's land drop rather than eating it, which is the
// whole reason the commander is built around cheap creatures.
//
// Until #654 the clause was omitted and declared: the pick-from-hand
// prompt existed (#552) but the hand-to-battlefield move did not.
func init() {
	Register(Spec{
		OracleID:        "ebf7ce9b-9e5e-4557-9e28-76556997f0ee",
		Name:            "Chulane, Teller of Tales",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureSpellCastByYou(ev, source, g)
			}, "Chulane, Teller of Tales — draw a card, then you may put a land from your hand onto the battlefield",
				Do(DrawCards{N: 1}, MayPutALandFromHand("Chulane, Teller of Tales"))),
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
