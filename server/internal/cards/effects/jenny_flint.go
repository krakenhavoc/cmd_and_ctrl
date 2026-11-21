package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jenny Flint — Legendary Creature — Human Detective {1}{U}{R}, 2/2 (Doctor Who):
//
//	"Partner with Madame Vastra (When this creature enters, target player may put
//	 Madame Vastra into their hand from their library, then shuffle.)
//	 First strike
//	 Training (Whenever this creature attacks with another creature with greater
//	 power, put a +1/+1 counter on this creature.)
//	 Whenever you sacrifice a Clue or Food, put a +1/+1 counter on another target
//	 creature you control."
//
// Partner with is the engine's entry trigger (CR 702.124j) and training the shared
// Training() trigger. The sacrifice trigger fires once per Clue or Food its controller
// sacrifices (the permanent as it last existed, so an animated or copied artifact is
// judged by what it was), chooses "another target creature you control" as it goes on
// the stack, and does nothing if that creature has gone by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d57adf99-972a-43fe-9ba2-2b01e673252f",
		Name:            "Jenny Flint",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Jenny Flint", "Madame Vastra"),
			Training("Jenny Flint — training: put a +1/+1 counter on it"),
			Targeting(
				On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && (c.HasSubtype("Clue") || c.HasSubtype("Food"))
				}, "Jenny Flint — put a +1/+1 counter on another target creature you control", b36CounterOnChosenAnimal),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
