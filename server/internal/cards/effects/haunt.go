package effects

import (
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

// hauntExileIt is the haunt ability's body. Package-level: nothing may be
// captured (undo resolves against a cloned game).
func hauntExileIt(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !SameGraveyardObject(z, item) {
		return nil
	}
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
	return g.CreatureItHauntsDiedForEffect(ev, source)
}
