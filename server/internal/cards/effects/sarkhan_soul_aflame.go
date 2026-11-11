package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan, Soul Aflame — Legendary Creature — Human Shaman {1}{U}{R}, 2/4:
//
//	"Dragon spells you cast cost {1} less to cast.
//	 Whenever a Dragon you control enters, you may have Sarkhan become a
//	 copy of it until end of turn, except its name is Sarkhan, Soul Aflame
//	 and it's legendary in addition to its other types."
//
// The discount is Dragonlord's Servant's. The trigger is Lazav's shape with
// a duration of until end of turn (CR 611.2): a "you may" with the Dragon as
// the shown subject, and a copy of the entered permanent's copiable values
// (CR 707.2) with the name and legendary exceptions (CR 707.9). While it is
// a copy Sarkhan has the Dragon's abilities and not his own, which is what
// the printed copy does; at end of turn he is himself again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbcae463-ba05-4cfe-9c25-13f2e64698c7",
		Name:         "Sarkhan, Soul Aflame",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Dragon spells you cast cost {1} less to cast.",
				YourSpell(), SpellOfCreatureType("Dragon")),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Dragon"))
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Sarkhan, Soul Aflame — become a copy of that Dragon until end of turn?", ShowSubject: true},
			Key:            "Sarkhan, Soul Aflame — becomes a copy of the Dragon until end of turn",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Sarkhan, Soul Aflame — becomes a copy of the Dragon until end of turn")
				item.Params.Object = game.ObjectRef{ID: ev.CardID}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BecomeCopy{
					Targets:  []uuid.UUID{ctx.Source()},
					Of:       item.Params.Object.ID,
					Duration: CopyUntilEndOfTurn,
					Except: func(v *game.PrintedValues) {
						v.SetName("Sarkhan, Soul Aflame")
						v.AddSupertype("Legendary")
					},
					Label: "Sarkhan, Soul Aflame — becomes a copy until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
