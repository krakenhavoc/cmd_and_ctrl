package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_of_exiled_with.go — "that creature becomes a copy of a creature
// card exiled with this Equipment" (#5399, Assimilation Aegis), and the
// "whenever this Equipment becomes attached to a creature" trigger that
// starts it.
//
// Three pieces, each the smallest addition to what was already there:
//
//   - The cards to pick from are CR 607.2a's linked set: the cards in
//     exile that a named ability of this same source object exiled
//     (Game.ExiledWithRowForEffect). The exile stamps the link when it is
//     ExileTarget.Linked or, for "exile … until this leaves the
//     battlefield", ExileUntil.Linked.
//   - One candidate is the copy; two or more are a pick (the ruling: you
//     choose one each time the ability resolves, not necessarily the same
//     one), a keyed choose_cards prompt whose continuation rebuilds the
//     duration from plain data, so a table waiting on it is a restore
//     point. None is no copy at all: the creature stays as it was.
//   - The copy is a BecomeCopy with any CopyDuration; Assimilation Aegis
//     takes CopyWhileSourceAttached, "for as long as this Equipment
//     remains attached to it" (CR 611.2b).
//
// The copy holds the card's copiable values as printed (CR 707.2,
// SingleFacedCopiableValuesOf; the ruling: nothing about the object the
// card was before it was exiled), read as the effect begins. It is a
// layer-1 copy effect (CR 613.2a) with a timestamp of its own, so
// counters and other layers' effects on the creature still apply.

// BecomeCopyOfExiledWith makes Target a copy of a card exiled with the
// resolving ability's source by the ability row Row (CR 607.2a), for
// Duration.
type BecomeCopyOfExiledWith struct {
	// Target is the permanent that becomes the copy.
	Target uuid.UUID

	// Slot and Row name the linked exiling ability: Slot is
	// game.AbilitySlotTriggered (the default) with the trigger's Key, or
	// game.AbilitySlotActivated with the row's Label.
	Slot, Row string

	// Match keeps the cards that may be copied ("a creature card": the
	// ruling refuses a noncreature card exiled the same way). Read on the
	// card where it sits in exile, before anything is asked; nil keeps
	// every linked card.
	Match func(c game.Card) bool

	// Duration is how long the copy lasts. A duration that cannot start
	// as the ability resolves copies nothing and asks nothing
	// (CR 611.2b).
	Duration CopyDuration

	// Question heads the pick when there are two or more candidates, and
	// Label names the copy effect in logs and tests.
	Question, Label string
}

// Apply lists the candidates and copies the one there is, or asks for
// one of two or more.
func (b BecomeCopyOfExiledWith) Apply(ctx *Context) error {
	src, ok := ctx.SourceRef()
	if !ok || b.Target == uuid.Nil {
		return nil
	}
	slot := b.Slot
	if slot == "" {
		slot = game.AbilitySlotTriggered
	}
	var cands []uuid.UUID
	for _, c := range ctx.Game.ExiledWithRowForEffect(game.PermissionCardRef(src), slot, b.Row) {
		if b.Match == nil || b.Match(*c) {
			cands = append(cands, c.InstanceID)
		}
	}
	switch len(cands) {
	case 0:
		return nil // the ruling: it remains just as it was
	case 1:
		return BecomeCopy{Targets: []uuid.UUID{b.Target}, Of: cands[0], Duration: b.Duration, Label: b.Label}.Apply(ctx)
	}
	// Don't ask a question whose answer can do nothing (CR 611.2b), nor
	// about a target that is no longer the object it was (#1432).
	targets := []uuid.UUID{b.Target}
	if ctx.isNewSourceObject(b.Target) {
		return nil
	}
	if _, ok := copyDurationFor(ctx.Game, b.Duration, ctx.Source(), ctx.Controller(), uuid.Nil, targets); !ok {
		return nil
	}
	host, ok := battlefieldObject(ctx.Game, b.Target)
	if !ok {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  b.Question,
		Cards:     cands,
		Min:       1,
		Max:       1,
		LiveZone:  game.ZoneExile,
		ThenKey:   becomeCopyOfExiledWithChosen,
		Carry:     targets,
		CarryWord: b.Label,
		CarryNums: []int{int(b.Duration), host.Epoch},
	})
	return nil
}

// battlefieldObject names the permanent `id` as an object: its ID and
// its zone-change count (CR 400.7). False when it is not on the
// battlefield.
//
// Caller must hold g.mu.
func battlefieldObject(g *game.Game, id uuid.UUID) (game.ObjectRef, bool) {
	if g.Battlefield == nil {
		return game.ObjectRef{}, false
	}
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
			return game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}, true
		}
	}
	return game.ObjectRef{}, false
}

// becomeCopyOfExiledWithChosen is BecomeCopyOfExiledWith's pick answered
// (#5399): Source is the ability's source, Carry[0] the permanent that
// becomes the copy, Nums its CopyDuration and that permanent's object
// epoch when it was asked, Word the effect's label. A permanent that is
// no longer that object, a duration that has ended meanwhile (the
// Equipment moved: CR 611.2b), or a pick withdrawn because the card left
// exile copies nothing. An on-disk key.
var becomeCopyOfExiledWithChosen = game.RegisterChooseCardsThen("choose-cards/become-copy-of-exiled-with", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || len(r.Nums) < 2 {
		return nil
	}
	host := r.Carry[0]
	if now, ok := battlefieldObject(g, host); !ok || now.Epoch != r.Nums[1] {
		return nil
	}
	targets := []uuid.UUID{host}
	d, ok := copyDurationFor(g, CopyDuration(r.Nums[0]), r.Source, r.Chooser, r.Picked[0], targets)
	if !ok {
		return nil
	}
	v, ok := g.CopiableValuesForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	label := r.Word
	if label == "" {
		label = "becomes a copy"
	}
	g.BecomeCopyForEffect(r.Source, r.Picked[0], targets, v, d, label)
	return nil
})

// ThisBecameAttachedToACreature is "Whenever this Equipment [or Aura]
// becomes attached to a creature" (#5399): an attach of this permanent
// onto a creature on the battlefield. A re-attach to the host it is
// already on does nothing (CR 701.3b, Event.AlreadyAttached), so nothing
// became attached and it does not trigger; an attach to a player or a
// noncreature permanent does not either. The host is the event's Target.
func ThisBecameAttachedToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttach || ev.CardID != source.InstanceID || ev.AlreadyAttached || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == ev.Target {
			return c.IsCreature()
		}
	}
	return false
}
