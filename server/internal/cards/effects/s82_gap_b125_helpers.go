package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b125_helpers.go — S82 set GAP, batch 125 (#5948): the Lord of the
// Rings Elf, Halfling and Treefolk rows. Append-only.

// gapB125PumpTrample gives one battlefield creature +p/+t and trample until
// end of turn: the shared body of Quickbeam, Mirkwood Channeler and Wose
// Pathfinder.
func gapB125PumpTrample(ctx *Context, target uuid.UUID, p, t int, label string) error {
	if target == uuid.Nil {
		return nil
	}
	if p != 0 || t != 0 {
		if err := (BoostUntilEOT{Target: target, Power: p, Toughness: t, Label: fmt.Sprintf("%s — %+d/%+d", label, p, t)}).Apply(ctx); err != nil {
			return err
		}
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample"}, Label: label + " — trample"}.Apply(ctx)
}

// gapB125TribeAttackAPlayer is the When of "whenever one or more <Tribe>s
// you control attack a player": a creature the source's controller controls,
// with the subtype read through the source's text, was declared attacking a
// player (not a planeswalker or battle).
func gapB125TribeAttackAPlayer(tribe string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !b16YouAttackedAPlayer(ev, source, g) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype(tribe))
	}
}

// gapB125ThisOrAnotherTreefolkEntered is "whenever this creature or another
// Treefolk you control enters" (the source's own entry, or another Treefolk
// entering under its controller's control).
func gapB125ThisOrAnotherTreefolkEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID ||
		b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Treefolk"))
}

// gapB125TwoOrMoreCreaturesDied is "as long as two or more creatures died
// under your control this turn": the turn tally's creature-death count
// for the source's controller.
func gapB125TwoOrMoreCreaturesDied(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.TurnTallyFor(you).CreaturesDied >= 2
}

// gapB125NoSharedCreatureType is the match of "a creature card that doesn't
// share a creature type with a creature you control": a creature card
// none of whose creature types (a changeling's are all of them) appears on
// a creature `you` control now (CR 205.3m, CR 702.73a).
func gapB125NoSharedCreatureType() CardPredicate {
	return func(g *game.Game, you uuid.UUID, c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		for _, mine := range g.BattlefieldCardsForEffect() {
			if mine.Controller != you || !mine.IsCreature() {
				continue
			}
			if game.SharesCreatureType(&c, &mine) {
				return false
			}
		}
		return true
	}
}

// gapB125CounteredCreatureTargetedByOpponent is "whenever a creature you
// control with a +1/+1 counter on it becomes the target of a spell or
// ability an opponent controls": the event names the creature and the
// targeting spell or ability's controller (its Actor), who is someone
// else.
func gapB125CounteredCreatureTargetedByOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller && c.Counters[game.CounterPlusOne] > 0
}
