package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murasa — Plane — Zendikar (Planechase Anthology, #4870):
//
//	"Whenever a nontoken creature enters, its controller may search their
//	 library for a basic land card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Whenever chaos ensues, target land becomes a 4/4 creature that's still a
//	 land."
//
// A plane (ADR 0216 §6). The first ability is the entering creature's
// controller's own "may", read when the creature entered (Genesis Chamber's
// shape), whoever that is. The chaos ability has no duration (the ruling):
// the land is a 4/4 creature, still a land, until the game ends, it leaves
// the battlefield, or another effect changes its types; it does not depend on
// Murasa staying face up. It is a layer 4 type add and a layer 7b set
// (CR 613.1d, 613.4b), pinned to the land so a new object is not it
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07cc253d-e23c-450c-bfaf-5cb6a3a588ce",
		Name:         "Murasa",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{murasaSearch(), murasaChaos()},
	})
}

func murasaSearch() game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := b26NontokenCreatureEntered(ev, g)
			return ok
		},
		Key: murasaSearchLabel,
		// The entering creature's controller is read as it entered and
		// carried as item.Params.Player.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			entered, _ := b26NontokenCreatureEntered(ev, g)
			item := game.NewTriggeredItem(source, murasaSearchLabel)
			item.Params.Player = entered.Controller
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if g.PlayerByIDForEffect(item.Params.Player) == nil {
				return nil
			}
			return SearchLibrary{
				Player:        item.Params.Player,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Optional:      true,
				Reason:        "Murasa — search for a basic land card",
			}.Apply(NewContext(g, item))
		},
	}
}

// murasaChaos is "target land becomes a 4/4 creature that's still a land",
// with no duration.
func murasaChaos() game.TriggeredAbility {
	return Targeting(WheneverChaosEnsues("Murasa — target land becomes a 4/4 creature that's still a land",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ScopedEffectFor{
				Target: id,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.SetBasePowerMod(4),
					game.SetBaseToughnessMod(4),
				},
				Duration: g.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Murasa — becomes a 4/4 creature that's still a land",
			}.Apply(ctx)
		}), TargetPermanent("target land", Land()))
}

const murasaSearchLabel = "Murasa — its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle"
