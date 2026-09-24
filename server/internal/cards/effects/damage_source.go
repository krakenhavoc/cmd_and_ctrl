package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damageSourceCharacteristics is a damage replacement event's SOURCE as
// a CR 614 matcher should read it — "a red source you control", "a red
// or artifact source you control" — and the same answer CR 702.16e
// protection reads (#1417).
//
// It is ev.SourceLKI, the snapshot every damage entry point takes when
// the event is created: the spell on the stack for a burn spell, the
// live permanent for a creature on the battlefield, and — the reason
// this helper exists — the permanent AS IT LAST EXISTED on the
// battlefield for one that has left (CR 608.2h). A creature an effect
// had turned red that dies and then deals its "when this dies" damage
// is a red source; the card in the graveyard is a new object with its
// printed colour (CR 400.7), and a stolen creature's graveyard card is
// back under its owner. g.LookupCardForEffect(ev.DamageSource) read
// that graveyard card, which is what these matchers did before #1417.
//
// The lookup survives only as the fallback for an event created
// without a snapshot, which no engine entry point does today.
//
// False when there is no source to read: a source-less event, or one
// the engine cannot find in any zone. Every caller treats that as "not
// a qualifying source", which errs weaker and never stronger.
func damageSourceCharacteristics(ev *game.ReplacementEvent, g *game.Game) (game.Characteristic, bool) {
	if ev == nil {
		return game.Characteristic{}, false
	}
	if ev.SourceLKI != nil {
		return *ev.SourceLKI, true
	}
	if ev.DamageSource == uuid.Nil {
		return game.Characteristic{}, false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	if !ok {
		return game.Characteristic{}, false
	}
	return *game.SourceCharacteristics(&c), true
}

// damageSourceIsRedControlledBy is "a red source <controller>
// controls", read through damageSourceCharacteristics. With
// orArtifact, "a red or artifact source" (Mechanized Warfare).
func damageSourceIsRedControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID, orArtifact bool) bool {
	ch, ok := damageSourceCharacteristics(ev, g)
	if !ok || ch.Controller != controller {
		return false
	}
	return slices.Contains(ch.Colors, "R") || (orArtifact && slices.Contains(ch.Types, "Artifact"))
}
