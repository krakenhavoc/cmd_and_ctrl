package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhuk, Hexgold Nabber — Legendary Creature — Goblin Rebel {2}{R}, 2/2
// (March of the Machine: Phyrexia: All Will Be One Commander):
//
//	"Trample, haste
//	 Whenever an equipped creature you control other than Rhuk attacks or
//	 dies, you may attach all Equipment attached to that creature to Rhuk."
//
// One ability watching two events (OnAny). "Equipped" is judged as the
// event happens: for an attack the creature is on the battlefield with an
// Equipment attached (CR 301.5); for a death the Equipment is still attached
// when the trigger is harvested, because the CR 704.5n unattach is a
// state-based action that has not run yet, and the creature is the
// controller's as it last existed (CR 603.10a). The Equipment to move are
// read as the ability resolves for a creature that is still the same object
// on the battlefield (an attacker), and from the list the trigger wrote down
// when the creature has gone (ObjectSnapshot.Attached), so a creature that
// died hands over the Equipment it died wearing even though they have since
// fallen off. Each is attached to Rhuk only if it is still the same
// Equipment on the battlefield; Rhuk himself must still be the same object
// on the battlefield, and the creature that died is not Rhuk (CR 701.3b).
// "You may" is asked as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06607cba-76ba-4145-a5b5-24d3428142ac",
		Name:            "Rhuk, Hexgold Nabber",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventAttack, game.EventLTB}, rhukEquippedCreatureAttackedOrDied,
				"Rhuk, Hexgold Nabber — attach all Equipment attached to that creature to Rhuk",
				rhukTakeTheEquipment),
				"Rhuk, Hexgold Nabber — attach all Equipment attached to that creature to Rhuk?"),
		},
	})
}

// rhukEquippedCreatureAttackedOrDied is the trigger condition.
func rhukEquippedCreatureAttackedOrDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	switch ev.Kind {
	case game.EventAttack:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && attackDeclaredByYou(ev, source.Controller) && c.Controller == source.Controller &&
			isEquippedCreature(g, c)
	case game.EventLTB:
		dead, ok := diedCreature(ev, g)
		return ok && leftUnderControlOf(ev, dead) == source.Controller && rhukWasEquipped(g, ev.CardID)
	}
	return false
}

// rhukWasEquipped reports whether an Equipment is attached to the permanent
// `id` right now.
func rhukWasEquipped(g *game.Game, id uuid.UUID) bool {
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(id) && a.HasSubtype("Equipment") {
			return true
		}
	}
	return false
}

// rhukTakeTheEquipment attaches every Equipment that was attached to the
// triggering creature to Rhuk.
func rhukTakeTheEquipment(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	creature := item.Trigger.Event.CardID
	var equipment []uuid.UUID
	if o := item.Trigger.Object; o != nil && onBattlefield(g, creature) && g.ObjectEpochForEffect(creature) == o.Epoch {
		// Still the same permanent: its Equipment as of now.
		for _, a := range g.BattlefieldCardsForEffect() {
			if a.IsAttachedTo(creature) && a.HasSubtype("Equipment") {
				equipment = append(equipment, a.InstanceID)
			}
		}
	} else if o != nil {
		// It has gone: the Equipment it had when the trigger was written down.
		for _, ref := range o.Attached {
			if onBattlefield(g, ref.ID) && g.ObjectEpochForEffect(ref.ID) == ref.Epoch {
				if a, ok := g.LookupCardForEffect(ref.ID); ok && a.HasSubtype("Equipment") {
					equipment = append(equipment, ref.ID)
				}
			}
		}
	}
	for _, id := range equipment {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}); err != nil {
			return err
		}
	}
	return nil
}
