package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d0ade00d-a496-441d-9b7e-7dc033d3292c",
		Name:         "Titania, Protector of Argoth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Key:       "Titania — return a land card from your graveyard to the battlefield",
				Targets:   effects.TargetCardInGraveyard("target land card in your graveyard", effects.YouOwn(), effects.Land()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.ReturnFromGraveyard{
						Target: item.Targets[0].ID,
						Dest:   game.ZoneBattlefield,
					}.Apply(effects.NewContext(g, item))
				},
			},
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10LandYouControlDied(ev, source, g)
			}, "Titania — create a 5/3 green Elemental", effects.Do(effects.CreateToken{Template: effects.TokenCard("5/3 green Elemental"), N: 1})),
		},
	})
}
