package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Progenitor's Icon — {3} Artifact (Marvel Super Heroes, #4455):
//
//	"As this artifact enters, choose a creature type.
//	 {T}: Add one mana of any color.
//	 {T}: The next spell of the chosen type you cast this turn can be cast as
//	 though it had flash."
//
// The type is named as it enters (CR 614.12, Cavern of Souls's prompt). The
// second ability is a one-use "next spell" flash promise (Apex Observatory's
// shape) whose filter is the type the Icon named, read when the ability
// resolves: a spell of the chosen type is any card with that creature type, so a
// changeling counts. The promise ends with the turn, and is spent by the next
// matching spell only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "747f6d0f-3c10-4fef-b7ae-f578c8964ab5",
		Name:         "Progenitor's Icon",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Progenitor's Icon"),
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: The next spell of the chosen type you cast this turn can be cast as though it had flash.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  progenitorsIconPromise,
		}},
	})
}

func progenitorsIconPromise(g *game.Game, item *game.StackItem) error {
	chosen := g.NamedTribeOf(item.SourceCardID)
	if chosen == "" {
		return nil
	}
	return GrantNextSpellPromise{From: "Progenitor's Icon", Promise: game.NextSpellPromise{
		Filter: game.PermissionFilter{CreatureType: chosen},
		Flash:  true,
		Text:   "The next " + chosen + " spell you cast this turn can be cast as though it had flash.",
	}}.Apply(NewContext(g, item))
}
