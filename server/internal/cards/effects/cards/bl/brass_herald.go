package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brass Herald — Artifact Creature — Golem {6}, 2/2 (Commander Legends,
// #5507):
//
//	"As this creature enters, choose a creature type.
//	 When this creature enters, reveal the top four cards of your library.
//	 Put all creature cards of the chosen type revealed this way into your
//	 hand and the rest on the bottom of your library in any order.
//	 Creatures of the chosen type get +1/+1."
//
// The type is named as it enters (Patchwork Banner's CR 614.12 prompt,
// stored on the permanent) and read by both clauses. The anthem is
// Coat of Arms' scope, not "you control": every creature of the type
// gets +1/+1, whoever controls it, the Herald included if it is of the
// type, read through the layered subtypes (a changeling counts). The
// reveal takes the matching creature cards by their printed subtypes in
// the library and puts the rest on the bottom in an order the controller
// chooses (Jace, Architect of Thought's ordering prompt).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea9b2e3d-2de4-4cde-b233-6d0ba5797c20",
		Name:         "Brass Herald",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Brass Herald"),
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Brass Herald — reveal the top four cards of your library; put all creature cards of the chosen type into your hand and the rest on the bottom of your library in any order",
				brassHeraldReveal),
		},
	})
}

func brassHeraldReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	tribe := g.NamedTribeOf(item.SourceCardID)
	revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 4,
		"Brass Herald — reveal the top four cards of your library")
	var bottom []uuid.UUID
	for _, id := range revealed {
		c, ok := g.LookupCardForEffect(id)
		if ok && tribe != "" && c.IsCreature() && c.HasSubtype(tribe) {
			if err := g.BounceToHandForEffect(id); err != nil {
				return err
			}
			continue
		}
		bottom = append(bottom, id)
	}
	if len(bottom) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Cards:     bottom,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Brass Herald — put the rest on the bottom of your library in any order",
	})
}
