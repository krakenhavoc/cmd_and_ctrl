package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herald's Horn — Artifact {3}:
//
//	"As this artifact enters, choose a creature type.
//	 Creature spells you cast of the chosen type cost {1} less to
//	 cast.
//	 At the beginning of your upkeep, look at the top card of your
//	 library. If it's a creature card of the chosen type, you may
//	 reveal it and put it into your hand."
//
// Three blockers on the batch-01 triage — cost modification, library
// top, choose a type — and all three are vocabulary now. The chosen
// type is `ChooseCreatureTypeAsEnters` landing on Card.NamedTribe; the
// discount is a `CostModifier` whose predicate reads that type off
// `q.Source`; and the upkeep is a look at the top card with a "you
// may" over it.
//
// # The discount
//
// It is a battlefield modifier (Spec.CostModifiers), not a self
// modifier, because it prices OTHER spells. The reduction spends
// against generic mana only and stops at zero, so a {G} one-drop Elf
// still costs {G} — that rule lives in the engine and no card file
// can get it wrong. Two Horns naming Elf stack.
//
// The predicate reads the SPELL's subtypes, which are printed
// characteristics, and a changeling passes for every type (CR
// 702.73a). With no type named yet nothing matches, which is the
// weaker direction and is the rule every chosen-type card follows.
//
// # The upkeep
//
// "Look at" is private — only the controller sees the card — and the
// reveal happens only on a yes, which is why it is inside the
// branch rather than before the question. A top card that is not a
// creature of the named type asks nothing at all: the card is looked
// at and left there, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c02c5547-b9c9-4b2d-9d12-e87bfba8f2d2",
		Name:         "Herald's Horn",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Herald's Horn"),
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Creature spells you cast of the chosen type cost {1} less to cast.",
				YourSpell(), CreatureSpell(), SpellOfTheSourcesChosenType()),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Herald's Horn — look at the top card of your library", heraldsHornUpkeep),
		},
	})
}

// heraldsHornUpkeep is the upkeep clause: look at the top card, and
// if it is a creature card of the named type offer to reveal it and
// take it.
func heraldsHornUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	tribe := g.NamedTribeOf(item.SourceCardID)
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if tribe == "" || len(looked) == 0 {
		return nil
	}
	top := looked[0]
	c, ok := g.LookupCardForEffect(top)
	if !ok || !c.IsCreature() || !c.HasSubtype(tribe) {
		return nil
	}
	player := item.Controller
	source := item.SourceCardID
	return MayChoice{
		Player:   player,
		Question: "Herald's Horn — reveal the top card and put it into your hand?",
		OnYes: func(ctx *Context) error {
			ctx.Game.RevealTopOfLibraryForEffect(player, source, 1, "Herald's Horn — reveal the top card")
			return ctx.Game.BounceToHandForEffect(top)
		},
	}.Apply(ctx)
}
