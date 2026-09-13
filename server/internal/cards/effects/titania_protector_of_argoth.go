package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titania, Protector of Argoth — Legendary Creature — Elemental
// {3}{G}{G}, 5/3 (EDHREC rank 1154):
//
//	"When Titania enters, return target land card from your graveyard
//	 to the battlefield.
//	 Whenever a land you control is put into a graveyard from the
//	 battlefield, create a 5/3 green Elemental creature token."
//
// The lands deck's payoff for every fetchland cracked and every land
// sacrificed. The ETB is Sun Titan's targeted graveyard return,
// narrowed to a land and mandatory; the land returns untapped, as
// printed, and its own enters-tapped replacement still applies. The
// second trigger reads the land post-move (diedCreature's posture),
// so a land sacrificed to a fetchland, to Zuran Orb, or destroyed by
// Strip Mine all count, and a land bounced or exiled does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0ade00d-a496-441d-9b7e-7dc033d3292c",
		Name:         "Titania, Protector of Argoth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetCardInGraveyard("target land card in your graveyard", YouOwn(), Land()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Titania — return a land card from your graveyard to the battlefield",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							return ReturnFromGraveyard{
								Target: item.Targets[0].ID,
								Dest:   game.ZoneBattlefield,
							}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b10LandYouControlDied(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Titania — create a 5/3 green Elemental",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{Controller: item.Controller, Template: b10GreenElemental53Token(), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
