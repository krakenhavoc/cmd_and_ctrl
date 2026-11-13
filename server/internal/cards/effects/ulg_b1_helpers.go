package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ulg_b1_helpers.go — the shapes Urza's Legacy, batch 1 (#5623) prints
// more than once, plus the two board readers its cards need. Append-only.

// ulgB1AnimateThisLand is a manland's "{cost}: This land becomes a P/T
// <colour> <subtype> creature [with <keywords>] until end of turn. It's
// still a land." (Forbidding Watchtower, Ghitu Encampment). It is Faerie
// Conclave's row: the land keeps its types and its mana ability, and an
// activation whose land has left, or come back as a new object (CR 400.7),
// does nothing.
func ulgB1AnimateThisLand(name, cost, label string, power, toughness int, subtype, colour string, keywords ...string) ActivatedAbility {
	desc := name + " — a " + strconv.Itoa(power) + "/" + strconv.Itoa(toughness) + " " + subtype + " creature until end of turn"
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerAnimate},
		Cost:    ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			return cmdr2B13Animate(NewContext(g, item), item.SourceCardID, desc,
				power, toughness, subtype, []string{colour}, keywords...)
		},
	}
}

// ulgB1AuraReturnsToHand is "When this Aura is put into a graveyard from
// the battlefield, return it to its owner's hand." (Sleeper's Guile,
// Sluggishness, Cessation): Despondency's trigger. It watches the Aura's
// own move from the battlefield to a graveyard, however it got there, and
// returns it only while it is still that graveyard object (CR 400.7).
func ulgB1AuraReturnsToHand(name string) game.TriggeredAbility {
	return WhenThisDies(name+" — return it to its owner's hand", returnThisCardFromYourGraveyardToHand)
}

// ulgB1DiesReturnToHand is "When this creature dies, return it to its
// owner's hand." (Shivan Phoenix, Weatherseed Treefolk): Immortal
// Phoenix's trigger.
func ulgB1DiesReturnToHand(name string) game.TriggeredAbility {
	return WhenThisDies(name+" — return it to its owner's hand", returnSelfFromGraveyardToHand)
}

// ulgB1ReturnAtEachEndStep is "At the beginning of the end step, return
// this creature to its owner's hand." (Viashino Sandscout, Viashino
// Cutthroat): Viashino Sandstalker's trigger, at every player's end step.
// BounceToHand's object guard leaves a creature that has already left
// alone.
func ulgB1ReturnAtEachEndStep(name string) game.TriggeredAbility {
	return AtEachStep(game.StepEnd, name+" — return it to its owner's hand", func(g *game.Game, item *game.StackItem) error {
		return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})
}

// ulgB1SacrificeToShrink is the Phyrexian Carriers' "{T}, Sacrifice this
// creature: Target creature gets -N/-N until end of turn." (Phyrexian
// Debaser, Defiler and Denouncer). Both are costs, paid as it is
// activated (CR 602.2b); a target gone in response gets nothing
// (CR 608.2b).
func ulgB1SacrificeToShrink(name string, n int) ActivatedAbility {
	pt := "-" + strconv.Itoa(n) + "/-" + strconv.Itoa(n)
	return ActivatedAbility{
		Label:   "{T}, Sacrifice this creature: Target creature gets " + pt + " until end of turn.",
		Purpose: ForTargets(HarmUntilEOT(0, -n, -n)),
		Cost:    Plus(TapCost(), SacrificeThis()),
		Targets: TargetCreature("target creature"),
		Effect:  boostTargetUntilEOT(name+" — "+pt, -n, -n),
	}
}

// cardsInAllHands is "the total number of cards in all players' hands"
// (Multani, Maro-Sorcerer). It runs inside a layer recompute, so it reads
// the seats directly.
func cardsInAllHands(g *game.Game) int {
	if g == nil {
		return 0
	}
	n := 0
	for _, p := range g.Seats {
		if p != nil && p.Hand != nil {
			n += p.Hand.Size()
		}
	}
	return n
}

// mountainsTheAurasControllerControls is Granite Grip's "each Mountain
// you control": the Aura's controller's Mountains, not the enchanted
// creature's controller's. "Mountain" is a word of the card's text
// (CR 612.2), so a text change counts another land type.
func mountainsTheAurasControllerControls(g *game.Game, source *game.Card) int {
	mountain := source.TextSubtype("Mountain")
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype(mountain) })
}
