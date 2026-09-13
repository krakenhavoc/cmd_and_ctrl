package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	b16AureliaDrawLabel   = "Aurelia, the Law Above — you draw a card (three or more attackers)"
	b16AureliaDamageLabel = "Aurelia, the Law Above — 3 damage to each opponent and you gain 3 life (five or more attackers)"
)

// Aurelia, the Law Above — Legendary Creature — Angel {3}{R}{W}, 4/4
// (EDHREC rank 1777):
//
//	"Flying, vigilance, haste
//	 Whenever a player attacks with three or more creatures, you draw
//	 a card.
//	 Whenever a player attacks with five or more creatures, Aurelia
//	 deals 3 damage to each of your opponents and you gain 3 life."
//
// The go-wide payoff that rewards everyone's alpha strikes. Three
// keywords on PrintedKeywords; two attack triggers on
// b16PlayerAttackedWithAtLeast, which counts the attacking creatures
// the declaring player controls after each declaration and fires
// each ability once per declaration — the engine emits EventAttack
// per creature, so the count reaching the threshold is what
// triggers, and the per-label dedup (queued, on the stack, or already
// fired this turn) is what keeps a six-creature attack from drawing
// four cards. "A player" is any player, Aurelia's controller
// included, and the damage goes to Aurelia's controller's opponents
// whoever attacked, as printed.
//
// Declared weaker than printed: with an extra combat in the same
// turn the abilities would not fire again. The engine has no extra
// combats, so nothing reaches it today.
func init() {
	Register(Spec{
		OracleID:        "2a800427-ff8c-4b3c-baee-85211b70656d",
		Name:            "Aurelia, the Law Above",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b16PlayerAttackedWithAtLeast(ev, source, g, 3, b16AureliaDrawLabel)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b16AureliaDrawLabel,
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b16PlayerAttackedWithAtLeast(ev, source, g, 5, b16AureliaDamageLabel)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b16AureliaDamageLabel,
						func(g *game.Game, item *game.StackItem) error {
							if err := damageToEachOpponent(g, item, 3); err != nil {
								return err
							}
							return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
