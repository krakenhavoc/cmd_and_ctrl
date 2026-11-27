package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b508_helpers.go — Shards of Alara, batch 508 (#5588): the shared
// bodies of the cycling cycle (the "Resounding" cards) and of the Heralds.

// OldB508CycleTrigger is the "When you cycle this card, <effect>" half of a
// Resounding card: it fires from the graveyard, where the card already is
// (CR 702.29c), one trigger whose target is chosen as it goes on the stack and
// judged again as it resolves (CR 603.3d, 608.2b).
func OldB508CycleTrigger(label string, spec *game.TargetSpec, effect Effect) game.TriggeredAbility {
	return Targeting(InGraveyard(On(game.EventCycle, Self, label, effect)), spec)
}

// OldB508CreatureOfColor is one entry of a Herald's sacrifice cost, "a white
// creature": a creature of that colour, read through the colour word (CR 612.2).
func OldB508CreatureOfColor(label, color string) game.SacrificeKind {
	return game.SacrificeKind{Label: label, CardTypes: []string{"creature"}, Colors: []string{color}}
}

// OldB508HeraldAbility is the Herald's one activated ability: pay `mana`, tap,
// and sacrifice three creatures of the three named colours (one for each, so a
// creature of two of them fills only one, CR 601.2h); then search the library
// for a card with the Herald's named Dragon or Sphinx and put it onto the
// battlefield, then shuffle. A search that finds nothing still shuffles
// (CR 701.23b).
func OldB508HeraldAbility(label, mana, named string, kinds ...game.SacrificeKind) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 1},
		Cost:    Plus(ManaCost(mana), TapCost(), SacrificeEach(OldB508KindsLabel(kinds), kinds...)),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return game.CardNameMatches(c, named) },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    label,
			}.Apply(NewContext(g, item))
		},
	}
}

// OldB508KindsLabel joins the kinds' labels as printed: "a, b, and c".
func OldB508KindsLabel(kinds []game.SacrificeKind) string {
	out := ""
	for i, k := range kinds {
		switch {
		case i == 0:
			out = k.Label
		case i == len(kinds)-1:
			out += ", and " + k.Label
		default:
			out += ", " + k.Label
		}
	}
	return out
}
