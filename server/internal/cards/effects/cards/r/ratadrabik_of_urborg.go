package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ratadrabik of Urborg — Legendary Creature — Zombie Wizard {2}{W}{B}, 3/3
// (Dominaria United):
//
//	"Vigilance, ward {2}
//	 Other Zombies you control have vigilance.
//	 Whenever another legendary creature you control dies, create a token
//	 that's a copy of that creature, except it's not legendary and it's a
//	 2/2 black Zombie in addition to its other colors and types."
//
// The lord half is Hordewing Skaab's TribeFilter. The death trigger reads the
// creature as it last existed on the battlefield (CR 603.10a: legendary,
// under your control) and copies it from last-known information (CR 608.2h,
// TokenCopyTemplate), so a token or a creature that has since left its
// graveyard still gets its copy. The exception clause is part of the copy
// (CR 707.9b): not legendary, base power and toughness 2/2, black added to
// its colours and Zombie added to its types. Ratadrabik itself dying does
// nothing ("another"), and an opponent's legend does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3ab43f2-4f2b-4439-9254-0110ce443fdb",
		Name:            "Ratadrabik of Urborg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "ward"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Ratadrabik of Urborg — ward {2}"),
			effects.On(game.EventLTB, ratadrabikAnotherLegendYouControlDied,
				"Ratadrabik of Urborg — create a token that's a copy of that creature, except it's not legendary and it's a 2/2 black Zombie in addition to its other colors and types",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.Trigger.Event.CardID,
						N:          1,
						Except:     ratadrabikExceptionFor(ctx.TextReader()),
					}.Apply(ctx)
				}),
		},
	})
}

// ratadrabikAnotherLegendYouControlDied is the death trigger's condition:
// another creature that was legendary and under the trigger's controller's
// control as it last existed on the battlefield went to a graveyard.
func ratadrabikAnotherLegendYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.LeftAsType(ev, dead, "Creature") && effects.LeftAsSupertype(ev, dead, "Legendary") &&
		effects.LeftUnderControlOf(ev, dead) == source.Controller
}

// ratadrabikException is "except it's not legendary and it's a 2/2 black
// Zombie in addition to its other colors and types" (CR 707.9b).
//
// The colour and the creature type are read through `reader`, Ratadrabik, the
// object whose text it is (CR 612.2).
func ratadrabikExceptionFor(reader *game.Card) func(t *game.Card) {
	return func(t *game.Card) { ratadrabikExceptionAs(t, reader.TextColor("B"), reader.TextSubtype("Zombie")) }
}

func ratadrabikExceptionAs(t *game.Card, black, zombie string) {
	t.TypeLine = effects.AddedSubtypeTypeLine(effects.RemoveLegendaryFromTypeLine(t.TypeLine), zombie)
	t.Power = 2
	t.Toughness = 2
	t.VariableToughness = false
	t.PrintedPTKnown = true
	colors := append([]string(nil), t.Colors...)
	for _, c := range colors {
		if c == black {
			t.SetCopyExceptionColors(colors...)
			return
		}
	}
	t.SetCopyExceptionColors(append(colors, black)...)
}
