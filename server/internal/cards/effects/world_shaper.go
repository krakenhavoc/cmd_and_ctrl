package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// World Shaper — Creature — Merfolk Shaman {3}{G}, 3/3 (EDHREC rank
// 1480):
//
//	"Whenever this creature attacks, you may mill three cards.
//	 When this creature dies, return all land cards from your
//	 graveyard to the battlefield tapped."
//
// The lands deck's self-mill on a body that pays out when it dies.
// The attack half is an optional trigger (the yes/no prompt, then
// the mill); the death half is Splendid Reclamation's body — every
// land card in the controller's graveyard returns under its owner's
// control.
//
// Sandbox simplification, declared, shared with Splendid Reclamation:
// the returned lands enter untapped and are tapped a beat later
// inside the same resolution, because the graveyard-return primitive
// has no tapped flag. They are tapped before any player gets
// priority; a "whenever a land enters untapped" watcher is the one
// thing that could tell.
func init() {
	Register(Spec{
		OracleID:     "3c075bb6-1831-4521-bd8d-4ed2825ae796",
		Name:         "World Shaper",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The returned lands enter untapped and are tapped a moment later, before anyone can act."},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: "World Shaper — mill three cards?",
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "World Shaper — mill three cards",
						func(g *game.Game, item *game.StackItem) error {
							return MillCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "World Shaper — return all land cards from your graveyard tapped",
						func(g *game.Game, item *game.StackItem) error {
							return b10ReturnAllLandCardsFromGraveyardTapped(NewContext(g, item), item.Controller)
						})
				},
			},
		},
	})
}
