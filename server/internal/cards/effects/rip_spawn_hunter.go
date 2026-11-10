package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rip, Spawn Hunter — Legendary Creature — Human Survivor {2}{G}{W}, 4/4
// (Duskmourn: House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if Rip is
//	 tapped, reveal the top X cards of your library, where X is its power.
//	 Put any number of creature and/or Vehicle cards with different powers
//	 from among them into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Survival (survival.go) is the trigger and its intervening if. X is Rip's
// power as it is now, or as it last existed once it has left (the
// rulings), read once as the ability resolves. The pick is a keyed
// choose_cards prompt with a registered set rule, "different powers"
// (ripDifferentPowers): a function of the picked cards alone, so a table
// waiting on it is a restore point and a bot is never offered a set the
// rule would refuse. A creature or Vehicle card is read in the library,
// where its power is the printed one; a card with no power (a Vehicle's
// is printed, a noncreature is not a candidate) never qualifies.
func init() {
	Register(Spec{
		OracleID:     "c027f369-e445-46a8-94ad-93396321295a",
		Name:         "Rip, Spawn Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("Rip, Spawn Hunter — reveal the top X cards (X is its power); put any number of creature and/or Vehicle cards with different powers into your hand; the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					x := info.Power
					if x <= 0 {
						return nil
					}
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, x,
						"Rip, Spawn Hunter — reveal the top X cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					var candidates []uuid.UUID
					for _, id := range revealed {
						if c, ok := g.LookupCardForEffect(id); ok && ripIsCreatureOrVehicle(c) {
							candidates = append(candidates, id)
						}
					}
					if len(candidates) == 0 {
						return g.PutOnBottomInRandomOrderForEffect(item.Controller, game.ZoneLibrary, revealed)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    item.Controller,
						FromPlayer: item.Controller,
						Source:     item.SourceCardID,
						Question:   "Rip, Spawn Hunter — put any number of creature and/or Vehicle cards with different powers into your hand",
						Cards:      candidates,
						Min:        0,
						Max:        len(candidates),
						LiveZone:   game.ZoneLibrary,
						ThenKey:    ripPicked,
						Carry:      revealed,
						Asked:      game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

func ripIsCreatureOrVehicle(c game.Card) bool {
	return c.IsCreature() || c.HasSubtype("Vehicle")
}

// ripDifferentPowers is "with different powers": no two picked cards share
// a power. On-disk behaviour of ripPicked; never captures game state.
func ripDifferentPowers(picked []game.Card) bool {
	seen := map[int]bool{}
	for _, c := range picked {
		p := c.PowerForComparison()
		if seen[p] {
			return false
		}
		seen[p] = true
	}
	return true
}

// ripPicked moves the chosen cards to the hand and the rest to the bottom of
// the library in a random order. An on-disk key.
var ripPicked = askedPickRule("choose-cards/rip-spawn-hunter-different-powers", ripDifferentPowers,
	func(ctx *Context, r game.CardsChosen) error {
		g := ctx.Game
		for _, id := range r.Picked {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
				continue
			}
			if err := g.BounceToHandForEffect(id); err != nil {
				return err
			}
		}
		return g.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, cardsStillInALibrary(g, r.Carry))
	})
