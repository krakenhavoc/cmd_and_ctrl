package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bard Class — {R}{G} Enchantment — Class (Adventures in the Forgotten Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Legendary creatures you control enter with an additional +1/+1 counter on
//	 them.
//	 {R}{G}: Level 2
//	 Legendary spells you cast cost {R}{G} less to cast. This effect reduces
//	 only the amount of colored mana you pay.
//	 {3}{R}{G}: Level 3
//	 Whenever you cast a legendary spell, exile the top two cards of your
//	 library. You may play them this turn."
//
// Level 1 is Gev's CR 614.12 entry replacement over the creatures entering under
// the Class's controller, narrowed to legendary ones; it has no gate (CR 716.3:
// the top section is always on). Level 2 is a colours-only reduction
// (CostModifier.UnitColorsOnly, the Defiler cycle's shape): its unit is {R}{G},
// which takes one red and one green mana off the cost where the cost has them
// and NOTHING off the generic part or another colour (CR 118.7). The spell must
// be legendary, read off the spell as announced, and cast by the Class's
// controller. Level 3 is a cast trigger gated at level 3 that exiles two cards
// with a play-this-turn permission over both, no limit on how many are played.
//
// No simplification.
func init() {
	unit, err := game.ParseCost("{R}{G}")
	if err != nil {
		panic("effects.BardClass: " + err.Error())
	}
	reduction := CostsLess(1, "Legendary spells you cast cost {R}{G} less to cast. This effect reduces only the amount of colored mana you pay.",
		YourSpell(), legendarySpell())
	reduction.Unit = &unit
	reduction.UnitColorsOnly = true
	reduction.ActiveWhen = Level(2)
	Register(Spec{
		OracleID:     "5bbc3ad0-4865-43f0-8baf-e7f4af5db656",
		Name:         "Bard Class",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{R}{G}")),
			LevelUp(3, ManaCost("{3}{R}{G}")),
		},
		Replacements:  []game.ReplacementEffect{bardClassLegendaryEntersWithCounter()},
		CostModifiers: []game.CostModifier{reduction},
		Triggered: []game.TriggeredAbility{
			AtLevel(3, WheneverYouCast(Legendary(),
				"Bard Class — exile the top two cards of your library; you may play them this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 2,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// bardClassLegendaryEntersWithCounter is "legendary creatures you control enter
// with an additional +1/+1 counter on them".
func bardClassLegendaryEntersWithCounter() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.IsLegendary() && entering.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Bard Class: legendary creatures you control enter with an additional +1/+1 counter on them",
	}
}
