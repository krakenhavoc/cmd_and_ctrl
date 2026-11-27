package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodbond March — Enchantment {2}{B}{G} (Ravnica: City of Guilds):
//
//	"Whenever a player casts a creature spell, each player returns all cards
//	 with the same name as that spell from their graveyard to the battlefield."
//
// One trigger over every player's creature spells (a creature as it sits on the
// stack). The spell's name is read as the trigger goes on the stack and carried
// on the item (Params.Name), so a spell that was countered, or whose name
// changed, in response still names the cards: "that spell" is the one cast
// (CR 603.10, last known information). A face-down spell has no name
// (CR 708.2a) and triggers nothing worth returning. As the ability resolves
// each seat, in turn order, returns every card in its own graveyard that has
// that name (CR 201.2) to the battlefield under its owner's control, since the
// text has no "under your control" (CR 400.3). The spell itself is still on
// the stack and is not in any graveyard.
//
// No simplification.
func init() {
	const label = "Bloodbond March — each player returns all cards with the same name as that spell from their graveyard to the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "fc17d8dd-887f-405e-a195-777aa3da36f5",
		Name:         "Bloodbond March",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsCreature()
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				if spell, ok := g.LookupCardForEffect(ev.CardID); ok && !spell.FaceDown {
					item.Params.Name = spell.Name
				}
				return item
			},
			Effect: BloodbondMarchReturn,
		}},
	})
}

// BloodbondMarchReturn is the resolution: every seat returns the cards in its
// graveyard named Params.Name to the battlefield.
func BloodbondMarchReturn(g *game.Game, item *game.StackItem) error {
	name := item.Params.Name
	if name == "" {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, p := range effects.TablePlayers(ctx) {
		ids := effects.GyCardIDs(g, p, func(c game.Card) bool { return game.CardNameMatches(c, name) })
		for _, id := range ids {
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
