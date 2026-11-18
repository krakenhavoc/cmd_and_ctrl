package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gorex, the Tombshell — Legendary Creature — Zombie Turtle {6}{B}{B}, 4/4
// (Innistrad: Midnight Hunt, #5895):
//
//	"As an additional cost to cast this spell, you may exile any number of
//	 creature cards from your graveyard. This spell costs {2} less to cast
//	 for each card exiled this way.
//	 Deathtouch
//	 Whenever Gorex attacks or dies, choose a card at random exiled with
//	 Gorex and put that card into its owner's hand."
//
// The cost is ExileAnyNumberFromGraveyardLinkedCost: the caster names any
// number of creature cards in their graveyard, none included, and each one
// takes {2} off the generic part (never below {B}{B}). The cards are linked
// to the spell and then to the permanent it becomes (CR 607.2q), so a
// countered Gorex links nothing, and a second Gorex has its own cards.
// The trigger is one ability with two conditions: each time, one linked
// card still in exile goes to its owner's hand, at random.
func init() {
	Register(Spec{
		OracleID:        "04238d49-941e-4298-8ef6-8e8e723f4034",
		Name:            "Gorex, the Tombshell",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		AdditionalCost: func() *game.AdditionalCost {
			ac := ExileAnyNumberFromGraveyardLinkedCost("X creature cards", func(c game.Card) bool { return c.IsCreature() })
			ac.Label = "You may exile any number of creature cards from your graveyard"
			return ac
		}(),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerExiledFromGraveyard("{2}", "Gorex, the Tombshell costs {2} less to cast for each card exiled this way."),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventLTB},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || cardDied(ev, source)
				},
				"Gorex, the Tombshell — put a card at random exiled with Gorex into its owner's hand",
				Do(ReturnRandomCardExiledWithThis{})),
		},
	})
}
