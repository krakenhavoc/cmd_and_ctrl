package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miirym, Sentinel Wyrm — Legendary Creature — Dragon Spirit
// {3}{G}{U}{R}:
//
//	"Flying, ward {2}
//	 Whenever another nontoken Dragon you control enters, create a
//	 token that's a copy of it, except the token isn't legendary."
//
// The copy is made from the Dragon that entered, read off the
// trigger's own event, so a Dragon that dies in response is still
// copied from its last-known values. The token is a real copy, so it
// brings the Dragon's oracle ID — and with it every ability the
// catalog has for it. Being a token, it does not trigger Miirym
// again, and a second Miirym watching the same event gets its own
// copy (each Miirym is an independent trigger).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "880fcf86-e6a9-482c-b91d-750f293127b2",
		Name:            "Miirym, Sentinel Wyrm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Miirym, Sentinel Wyrm — ward {2}"),
			On(game.EventETB, AnotherNontokenCreatureOfTypeEnteredUnderYourControl("Dragon"),
				"Miirym, Sentinel Wyrm — create a token that's a copy of that Dragon, except it isn't legendary",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.Trigger.Event.CardID,
						N:          1,
						Except: func(t *game.Card) {
							t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
