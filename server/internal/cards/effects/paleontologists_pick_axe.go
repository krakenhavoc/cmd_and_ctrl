package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paleontologist's Pick-Axe // Dinosaur Headdress — a transforming
// Equipment (#2709, ADR 0137 and its 2026-10-10 amendment):
//
//	Paleontologist's Pick-Axe — Artifact — Equipment {2}
//	  "Whenever equipped creature attacks, draw a card, then discard a
//	   card.
//	   Equip {1}
//	   Craft with one or more creatures {5}"
//	Dinosaur Headdress — Artifact — Equipment
//	  "When this Equipment enters, attach it to target creature you
//	   control.
//	   As this Equipment becomes attached to a creature, choose an
//	   exiled creature card used to craft this Equipment.
//	   Equipped creature is a copy of the last chosen card.
//	   Equip {2}"
//
// The Pick-Axe is Argentum Armor's attack trigger with a loot. The
// craft is an open count of creatures from either zone.
//
// The Headdress's choice is an AsAttached clause (CR 614.12's shape for
// an attach): the engine's one attach verb runs it every time the
// Headdress becomes attached to a creature it was not already on — the
// enters trigger, an equip, anything else — and the controller picks a
// creature card among its materials still in exile (CR 702.167c), the
// same one as last time if they like (the ruling). The pick becomes a
// CR 707.2 copy effect on that creature, holding the card's copiable
// values as printed (the rulings: exactly what is printed, X is 0), for
// as long as the Headdress stays attached to it
// (game.WhileSourceAttachedToPinned). Moving the Headdress ends that
// copy and makes a new choice for the new creature; unattaching it
// ends the copy. The copy keeps the values it took if the chosen card
// later leaves exile, as every duration copy does.
//
// No simplification.
const paleontologistsPickAxeOracleID = "9b0c6b77-69df-4330-86a7-8dc4666abd38"

func init() {
	Register(Spec{
		OracleID:     paleontologistsPickAxeOracleID,
		Name:         "Paleontologist's Pick-Axe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			Key: "Paleontologist's Pick-Axe — draw a card, then discard a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
			Craft("Craft with one or more creatures {5}", "{5}", CraftWithOneOrMore("creature")),
		},
	})

	Register(Spec{
		OracleID:     paleontologistsPickAxeOracleID + "#1",
		Name:         "Dinosaur Headdress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dinosaur Headdress — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		AsAttached: dinosaurHeaddressChoose,
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// dinosaurHeaddressChoose is the "as becomes attached" choice, and the
// copy effect it starts on the creature.
func dinosaurHeaddressChoose(card *game.Card, ctx *Context) error {
	host := card.AttachedTo.ID
	if card.AttachedTo.Kind != game.TargetCard || host == uuid.Nil {
		return nil
	}
	var creatures []uuid.UUID
	for _, c := range craftMaterialsOf(ctx.Game, card) {
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Dinosaur Headdress — choose an exiled creature card used to craft it; the equipped creature becomes a copy of it",
		Cards:    creatures,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  dinosaurHeaddressCopy,
		Carry:    []uuid.UUID{host},
	})
	return nil
}

// dinosaurHeaddressCopy is the exiled creature card the equipped creature
// becomes a copy of, keyed (#4253): the Headdress is the prompt's source and
// the creature it was attached to rides the prompt as the carried ID.
var dinosaurHeaddressCopy = game.RegisterChooseCardsThen("choose-cards/dinosaur-headdress-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	headdress, host := r.Source, r.Carry[0]
	d, ok := g.WhileAttachedToDuration(headdress, host)
	if !ok {
		return nil
	}
	v, ok := g.CopiableValuesForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	g.BecomeCopyForEffect(headdress, r.Picked[0], []uuid.UUID{host}, v, d, "Dinosaur Headdress")
	return nil
})
