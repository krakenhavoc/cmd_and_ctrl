package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elenda, the Dusk Rose — Legendary Creature — Vampire Knight
// {2}{W}{B}, 1/1:
//
//	"Lifelink
//	 Whenever another creature dies, put a +1/+1 counter on Elenda.
//	 When Elenda dies, create X 1/1 white Vampire creature tokens with
//	 lifelink, where X is Elenda's power."
//
// "Another creature" is anyone's creature, token or card
// (AnotherCreatureDied), and each death is its own trigger, so a wrath
// that kills five creatures with Elenda on the table gives her five
// counters if she survives it. A creature that dies in the same event as
// Elenda still triggers her (CR 603.10a looks back), but she is gone by
// resolution and a counter has no permanent to land on, as printed.
//
// X IS ELENDA'S LAST-KNOWN POWER (CR 603.10, 608.2h): the layer-applied
// power she had on the battlefield with every +1/+1 counter and anthem
// counted, not the printed 1 and not the card in the graveyard.
// b13LastKnownPower is that read, and it is taken when the trigger is
// built, so a later change cannot move it. X of zero or less makes
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce35f262-b93f-4409-8b33-26a7ed7abac8",
		Name:            "Elenda, the Dusk Rose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AnotherCreatureDied, "Elenda, the Dusk Rose — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Key: "Elenda, the Dusk Rose — a lifelink Vampire for each point of her power",
				Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Elenda, the Dusk Rose — a lifelink Vampire for each point of her power")
					item.Params.Amount = b13LastKnownPower(g, source.InstanceID, lki)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Params.Amount <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: b36WhiteVampireLifelinkToken(), N: item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
