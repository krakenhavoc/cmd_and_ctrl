package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clone Shell — Artifact Creature — Shapeshifter {5}, 2/2 (Scars of Mirrodin,
// S82 OLD batch 472, #5577):
//
//	"Imprint — When this creature enters, look at the top four cards of your
//	 library, exile one face down, then put the rest on the bottom of your
//	 library in any order.
//	 When this creature dies, turn the exiled card face up. If it's a creature
//	 card, put it onto the battlefield under your control."
//
// The imprint is the look-and-exile of Urianger Augurelt (the card goes to
// exile face down, seen only by its controller, linked to this object,
// CR 607.2a, CR 406.3) over four cards, with the exile required, then the rest
// ordered onto the bottom as Proteus Staff does. The dies trigger reads the
// exile as the Shell last existed (Prowling Geistcatcher's reading of its own
// leaving), turns the card face up for the table, and a creature card goes onto
// the battlefield under the Shell's controller as the ability resolves; any
// other card stays in exile face up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94ab7d13-bcd7-419f-9e68-bd8e7ef66d42",
		Name:         "Clone Shell",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Clone Shell — look at the top four cards of your library, exile one face down, then put the rest on the bottom in any order", cloneShellImprint),
			effects.WhenThisDies("Clone Shell — turn the exiled card face up; if it's a creature card, put it onto the battlefield under your control", cloneShellFlip),
		},
	})
}

func cloneShellImprint(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	looked := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 4)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Clone Shell — exile one face down",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    cloneShellPicked,
		Carry:      []uuid.UUID{ref.ID},
		CarryNums:  []int{ref.Epoch},
	})
	return nil
}

// cloneShellPicked is the pick answered: the card goes to exile face down,
// linked to the Shell object, and the other cards looked at go on the bottom
// in the order the controller picks. An on-disk key: never renamed, never
// reused.
var cloneShellPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b472-clone-shell-imprint", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
		src := game.PermissionCardRef{ID: r.Carry[0], Epoch: r.Nums[0]}
		if _, err := g.ExileHiddenLinkedForEffect(src, r.Chooser, picked); err != nil {
			return err
		}
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   r.Chooser,
		Source:    r.Source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    effects.LibraryOrderLabel(g, r.Source, "put the rest on the bottom of your library in any order"),
	})
})

func cloneShellFlip(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		g.TurnExiledFaceUpForEffect(item.SourceCardID, id, "Clone Shell — turn the exiled card face up")
		if c, found := g.LookupCardForEffect(id); found && c.IsCreature() {
			if err := (effects.ReturnFromExileTogether{Targets: []uuid.UUID{id}, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
