package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spelunking — Enchantment {2}{G} (EDHREC rank 740):
//
//	"When this enchantment enters, draw a card, then you may put a
//	 land card from your hand onto the battlefield. If you put a Cave
//	 onto the battlefield this way, you gain 4 life.
//	 Lands you control enter untapped."
//
// The static is the card: every tapped dual, Temple and bounce land
// you play enters untapped. It is a CR 614 replacement on the entering
// land's move that clears EntersTapped. When the land carries its own
// "enters tapped" replacement, both apply to the same event and CR
// 616.1 has the land's controller order them — choosing Spelunking's
// last is what makes the land untapped, exactly the paper interaction.
// A land with no such clause sees Spelunking's alone, which is a no-op
// with no prompt.
//
// Sandbox simplifications, both weaker than printed:
//
//   - The ETB's "you may put a land card from your hand onto the
//     battlefield" (and the Cave life rider) is not implemented — no
//     pick-from-hand prompt with a continuation exists, the Growth
//     Spiral gap. The ETB draws its card and stops.
//   - A land put onto the battlefield TAPPED by an effect (Cultivate,
//     Evolving Wilds, Riveteers Overlook — the SearchLibrary
//     TappedOnEntry flag) still enters tapped: that flag is applied
//     after the replacement pipeline and OR-ed with it. Only the
//     land's own enters-tapped clause is overridden.
func init() {
	Register(Spec{
		OracleID:     "2962fe4c-bf48-454b-8a6b-0f8253352ae8",
		Name:         "Spelunking",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The entry trigger only draws the card — it doesn't offer to put a land from your hand onto the battlefield, and the Cave life bonus never happens.",
			"A land an effect puts onto the battlefield tapped still enters tapped; only a land's own enters-tapped text is overridden.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Spelunking — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
					return false
				}
				if ev.Actor != src.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = false
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Spelunking: lands you control enter untapped",
		}},
	})
}
