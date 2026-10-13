package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neriv, Crackling Vanguard — Legendary Creature — Spirit Dragon
// {2}{R}{W}{B}, 4/4:
//
//	"Flying, deathtouch
//	 When Neriv enters, create two 1/1 red Goblin creature tokens.
//	 Whenever Neriv attacks, exile a number of cards from the top of
//	 your library equal to the number of differently named tokens you
//	 control. During any turn you attacked with a commander, you may
//	 play those cards."
//
// The count is read as the attack trigger resolves: the distinct NAMES
// among the tokens its controller controls then (two Goblins and a
// Treasure is two). Zero tokens exiles nothing.
//
// The permission is a conditional exile play (#3192, ADR 0066's second
// 2026-10-13 amendment): the cards stay playable for as long as they
// stay exiled, during any turn their holder attacked with a commander
// (game.CondAttackedWithCommander, read off the attacker as it was
// declared). Neriv is a commander in the decks that run her, but any
// commander attacking opens the cards, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8f827d7-4740-43c8-b494-c9e0479e0bf7",
		Name:            "Neriv, Crackling Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Neriv, Crackling Vanguard — create two 1/1 red Goblin creature tokens", Do(CreateToken{
				Template: RedGoblinToken(), N: 2,
			})),
			WheneverThisAttacks("Neriv, Crackling Vanguard — exile cards equal to the number of differently named tokens you control", func(g *game.Game, item *game.StackItem) error {
				names := map[string]bool{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && IsToken(c) {
						names[c.Name] = true
					}
				}
				if len(names) == 0 {
					return nil
				}
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: len(names),
					WhileExiled: true,
					Condition:   attackedCondition(game.CondAttackedWithCommander),
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
