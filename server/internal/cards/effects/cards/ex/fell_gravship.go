package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fell Gravship — Artifact — Spacecraft {2}{B}, 3/2 (Edge of Eternities,
// #4461):
//
//	"When this Spacecraft enters, mill three cards, then return a
//	 creature or Spacecraft card from your graveyard to your hand.
//	 Station
//	 8+ | Flying, lifelink"
//
// Carrion Cruiser's enters trigger with three cards and Spacecraft in
// place of Vehicle: the mill finishes first (its continuation), and the
// pick ranges over the whole graveyard, the milled cards included
// (CR 608.2c). The return is not targeted and not optional: a lone
// candidate returns without a question, two or more ask for a pick with
// a floor of one, and with none nothing happens. Station and the 8+ line
// are Susurian Dirgecraft's helpers: at eight or more charge counters it
// is a 3/2 artifact creature with flying and lifelink (CR 702.184).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a82be68-3b74-4dfc-9068-3abea61db709",
		Name:         "Fell Gravship",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fell Gravship — mill three cards, then return a creature or Spacecraft card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(fellGravshipMilled)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(8, 3, 2),
			effects.ThresholdKeywords(8, "flying", "lifelink"),
		),
	})
}

// fellGravshipMilled is the trigger's "then": once the mill has settled,
// a creature or Spacecraft card from the controller's graveyard goes to
// their hand. An on-disk key: never renamed, never reused.
var fellGravshipMilled = effects.CardTail("fell-gravship/after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	match := effects.Or(effects.Creature(), effects.HasSubtype("Spacecraft"))
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(ctx.Game, ctx.Controller(), c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 1 {
		return effects.ReturnFromGraveyard{Target: candidates[0], Dest: game.ZoneHand}.Apply(ctx)
	}
	return effects.TakeOneIntoYourHand(ctx, candidates,
		"Fell Gravship — return a creature or Spacecraft card from your graveyard to your hand", 1)
})
