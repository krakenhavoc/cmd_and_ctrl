package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// haunt.go — the card-side vocabulary of haunt (CR 702.55, #6083). The
// engine half, the link between an exiled card and the creature it haunts,
// is game/haunt.go.
//
// A haunt creature prints two triggered abilities:
//
//	Haunt (When this creature dies, exile it haunting target creature.)
//	When this creature enters or the creature it haunts dies, <effect>.
//
// The first is Haunt; the second is WhenThisEntersOrTheCreatureItHauntsDies,
// ONE ability with two trigger conditions (Blind Hunter, Orzhov Euthanist,
// Orzhov Pontiff). Its second condition is met in exile (CR 702.55c), so it
// declares the battlefield and exile both, the way an eminence trigger
// declares the battlefield and the command zone.
//
// A haunt instant or sorcery prints the other half of CR 702.55a:
//
//	Haunt (When this spell card is put into a graveyard after resolving,
//	exile it haunting target creature.)
//	When the creature this card haunts dies, <effect>.
//
// The first is HauntSpell, which triggers from the graveyard on the move
// the engine marks Event.Resolved (a countered spell does not haunt); the
// second is WhenTheCreatureThisCardHauntsDies, which triggers only from
// exile (Cry of Contrition, Benediction of Moons, Seize the Soul).

// Haunt is "Haunt (When this creature dies, exile it haunting target
// creature.)" — CR 702.55a. `name` is the card's name, for the stack label.
//
// The card is exiled only if it is still the object that died, in its
// owner's graveyard (CR 400.7, 608.2b), and the target is still legal:
// with no legal target the ability is removed from the stack (CR 603.3d)
// or does nothing (CR 608.2b), and the card stays in the graveyard.
func Haunt(name string) game.TriggeredAbility {
	t := Targeting(
		WhenThisDies(name+" — haunt: exile it haunting target creature", hauntExileIt),
		TargetCreature("target creature"),
	)
	t.Keyword = "haunt"
	return t
}

// HauntSpell is "Haunt (When this spell card is put into a graveyard after
// resolving, exile it haunting target creature.)" — CR 702.55a's second
// sentence. It watches from the graveyard (CR 113.6), where the card is
// as it triggers, for the EventZoneMove from the stack that the engine
// marks Resolved (CR 608.2n): a spell countered, or removed for want of
// legal targets (CR 608.2b), lands there unmarked and does not haunt, and
// a spell that went somewhere else (flashback, rebound, a replacement)
// was never put into a graveyard.
//
// The card is exiled only if it is still the object that landed (CR
// 400.7) and the target is still legal; otherwise it stays where it is.
func HauntSpell(name string) game.TriggeredAbility {
	t := Targeting(
		On(game.EventZoneMove, thisSpellResolvedIntoAGraveyard,
			name+" — haunt: exile it haunting target creature", hauntExileThisSpellCard),
		TargetCreature("target creature"),
	)
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	t.Keyword = "haunt"
	return t
}

// thisSpellResolvedIntoAGraveyard is HauntSpell's condition: this card
// was put into a graveyard from the stack as its own resolution ended.
func thisSpellResolvedIntoAGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Resolved && ev.CardID == source.InstanceID && ev.OldZone == game.ZoneStack && ev.NewZone == game.ZoneGraveyard
}

// hauntExileThisSpellCard is HauntSpell's body: the card must still be the
// graveyard object the ability triggered from, which is the object its
// item names (no move since, CR 400.7).
func hauntExileThisSpellCard(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameObjectStillHere(z, item) {
		return nil
	}
	return hauntTheTarget(g, item)
}

// sameObjectStillHere reports whether the card the item names is, in zone
// z, the very object the ability triggered from: its epoch has not moved
// since. An item with no stamp keeps the plain "is it here" answer, as
// SameGraveyardObject does.
func sameObjectStillHere(z *game.Zone, item *game.StackItem) bool {
	ref := item.SourceObject
	if ref.ID != item.SourceCardID || ref.ID == uuid.Nil {
		return true
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == ref.ID {
			return z.Cards[i].ObjectEpoch == ref.Epoch
		}
	}
	return false
}

// WhenTheCreatureThisCardHauntsDies is "When the creature this card haunts
// dies, <effect>" on a haunt instant or sorcery (CR 702.55b, 702.55c). It
// triggers only from exile, where HauntSpell put the card; its owner
// controls it (CR 108.4).
func WhenTheCreatureThisCardHauntsDies(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventLTB, theCreatureItHauntsDied, label, effect)
	t.Zones = []game.ZoneKind{game.ZoneExile}
	return t
}

// theCreatureItHauntsDied is the exile-only condition.
func theCreatureItHauntsDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.CreatureItHauntsDiedForEffect(ev, source)
}

// hauntExileIt is the haunt ability's body. Package-level: nothing may be
// captured (undo resolves against a cloned game).
func hauntExileIt(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !SameGraveyardObject(z, item) {
		return nil
	}
	return hauntTheTarget(g, item)
}

// hauntTheTarget exiles the item's source card haunting the item's target
// creature, if that target is still legal.
func hauntTheTarget(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		g.HauntForEffect(item.SourceCardID, t.ID)
		return nil
	}
	return nil
}

// WhenThisEntersOrTheCreatureItHauntsDies is "When this creature enters or
// the creature it haunts dies, <effect>": one ability, two conditions
// (CR 702.55b, 702.55c). On the battlefield it triggers on its own entry;
// in exile, where a haunt ability put it, on the death of the creature it
// haunts. The exiled card's owner controls the second (CR 108.4).
func WhenThisEntersOrTheCreatureItHauntsDies(label string, effect Effect) game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventETB, game.EventLTB}, enteredOrTheCreatureItHauntsDied, label, effect)
	t.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneExile}
	return t
}

// enteredOrTheCreatureItHauntsDied is the two conditions: this object
// entering the battlefield, or (in exile) the creature it haunts dying.
func enteredOrTheCreatureItHauntsDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return ev.CardID == source.InstanceID
	}
	return theCreatureItHauntsDied(ev, source, game.Characteristic{}, g)
}
