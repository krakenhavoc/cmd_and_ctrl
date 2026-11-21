package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapArtifactCreatureYouControlDied is "Nick Valentine or another artifact
// creature you control dies": the permanent that left to the graveyard was
// an artifact creature, under the source controller's control, as it last
// existed on the battlefield (CR 603.10a).
func GapArtifactCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && LeftAsType(ev, dead, "artifact") && LeftUnderControlOf(ev, dead) == source.Controller
}

// gapYourCreatureEnlisted is "a creature you control enlisted a creature":
// EventEnlist's Actor is the attacker's controller.
func GapYourCreatureEnlisted(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && ev.CardID != uuid.Nil && ev.Target != uuid.Nil
}

// gapAradeshResolve gives the attacker named by the triggering enlist event
// double strike until end of turn and draws a card if its power is now 4 or
// greater. An attacker that has left the battlefield does nothing.
func GapAradeshResolve(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !B15OnBattlefield(g, attacker) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (GrantKeywordUntilEOT{
		Target:   attacker,
		Keywords: []string{"double strike"},
		Label:    "Aradesh, the Founder — double strike until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(attacker); ok && c.CurrentPower() >= 4 {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}

// gapZombieOrMutantDiedWithOddPower is "a Zombie or Mutant you control dies,
// if its power was different from its base power": the creature as it last
// existed on the battlefield (CR 603.10a) was a Zombie or a Mutant under my
// control, and its power was not its base power (CR 208.4b).
func GapZombieOrMutantDiedWithOddPower(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok || LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	if !LeftAsSubtype(ev, dead, source.TextSubtype("Zombie")) && !LeftAsSubtype(ev, dead, source.TextSubtype("Mutant")) {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	return ok && info.Power != info.BasePower
}
