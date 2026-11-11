package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ishkanah, Grafwidow — Legendary Creature — Spider {4}{G}, 3/5 (Duskmourn
// Commander):
//
//	"Reach
//	 Delirium — When Ishkanah enters, if there are four or more card types
//	 among cards in your graveyard, create three 1/2 green Spider creature
//	 tokens with reach.
//	 {6}{B}: Target opponent loses 1 life for each Spider you control."
//
// The enters trigger is an intervening if (CR 603.4): checked as Ishkanah
// enters and again as the trigger resolves (b24CondDelirium, the caster's own
// graveyard). The drain counts the Spiders you control as the ability
// resolves, read through the layers, Ishkanah and the tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6f4861f4-1b3b-4dbc-b55b-196075d0a693",
		Name:            "Ishkanah, Grafwidow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(b24CondDelirium{}, On(game.EventETB, AllOf(Self, youHaveDeliriumWhen),
				"Ishkanah, Grafwidow — create three 1/2 green Spider creature tokens with reach",
				Do(CreateTokenAdvanced{Spec: Token(TokenCard("1/2 green Spider with reach")), N: 3}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}{B}: Target opponent loses 1 life for each Spider you control.",
			Cost:    ManaCost("{6}{B}"),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtypeWord("Spider") })
				if n == 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n)
					}
				}
				return nil
			},
		}},
	})
}
