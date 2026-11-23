package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phenomenon Investigators — Creature — Human Detective {2}{U}{B}, 3/4
// (Duskmourn Commander, #5368):
//
//	"As this creature enters, choose Believe or Doubt.
//	 • Believe — Whenever a nontoken creature you control dies, create a
//	   2/2 black Horror enchantment creature token.
//	 • Doubt — At the beginning of your end step, you may return a
//	   nonland permanent you own to your hand. If you do, draw a card."
//
// The choice is Windcrag Siege's (ChooseOptionAsEnters, each line gated
// by WhenChosen), so the unchosen line never triggers.
//
// Believe is Blight Mound's dies trigger over the printed Horror: a
// creature that left the battlefield for a graveyard under your control,
// read from its last-known information, and not a token — Phenomenon
// Investigators itself counts.
//
// Doubt does not target: it is a resolution-time choice over the nonland
// permanents you OWN (an opponent-controlled permanent you own is on
// offer), floor zero for "you may" (Ambrosia Whiteheart's pick). The draw
// is "if you do": it runs from the bounce's tail, only when the permanent
// actually reached your hand (a commander that went to the command zone
// instead draws nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc9088c4-6f6b-4a1a-90f1-794eb1e938c6",
		Name:         "Phenomenon Investigators",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Phenomenon Investigators", "Believe", "Doubt"),
		Triggered: []game.TriggeredAbility{
			effects.WhenChosen("Believe", effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.IsToken(dead)
			}, "Phenomenon Investigators — create a 2/2 black Horror enchantment creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Horror enchantment"), N: 1}))),
			effects.WhenChosen("Doubt", effects.AtYourEndStep(
				"Phenomenon Investigators — you may return a nonland permanent you own to your hand; if you do, draw a card",
				phenomenonInvestigatorsDoubt)),
		},
	})
}

// phenomenonInvestigatorsDoubt asks which nonland permanent the
// controller owns to return, if any.
func phenomenonInvestigatorsDoubt(g *game.Game, item *game.StackItem) error {
	return effects.ChoosePermanents{
		Question: "Phenomenon Investigators — you may return a nonland permanent you own to your hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Owner == of && !c.IsLand() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 1
		},
		Keyed: phenomenonInvestigatorsReturned,
	}.Apply(effects.NewContext(g, item))
}

// phenomenonInvestigatorsReturned bounces the chosen permanent, and draws
// from the bounce's tail. On-disk keys: never renamed, never reused.
var phenomenonInvestigatorsReturned = effects.RegisterPermanentsPick("prompt-run/phenomenon-investigators-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			return effects.BounceToHand{Target: id, ThenTail: ctx.Tail(phenomenonInvestigatorsDraw)}.Apply(ctx)
		}
		return nil
	},
})

var phenomenonInvestigatorsDraw = effects.CardTail("then/phenomenon-investigators-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
