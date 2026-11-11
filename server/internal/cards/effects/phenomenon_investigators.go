package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "cc9088c4-6f6b-4a1a-90f1-794eb1e938c6",
		Name:         "Phenomenon Investigators",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Phenomenon Investigators", "Believe", "Doubt"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Believe", On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead)
			}, "Phenomenon Investigators — create a 2/2 black Horror enchantment creature token",
				Do(CreateToken{Template: TokenCard("2/2 black Horror enchantment"), N: 1}))),
			WhenChosen("Doubt", AtYourEndStep(
				"Phenomenon Investigators — you may return a nonland permanent you own to your hand; if you do, draw a card",
				phenomenonInvestigatorsDoubt)),
		},
	})
}

// phenomenonInvestigatorsDoubt asks which nonland permanent the
// controller owns to return, if any.
func phenomenonInvestigatorsDoubt(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
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
	}.Apply(NewContext(g, item))
}

// phenomenonInvestigatorsReturned bounces the chosen permanent, and draws
// from the bounce's tail. On-disk keys: never renamed, never reused.
var phenomenonInvestigatorsReturned = RegisterPermanentsPick("prompt-run/phenomenon-investigators-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			return BounceToHand{Target: id, ThenTail: ctx.Tail(phenomenonInvestigatorsDraw)}.Apply(ctx)
		}
		return nil
	},
})

var phenomenonInvestigatorsDraw = CardTail("then/phenomenon-investigators-draw", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
})
