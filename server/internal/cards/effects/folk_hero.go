package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Folk Hero — {1}{W} Legendary Enchantment — Background:
//
//	"Commander creatures you own have "Whenever you cast a spell that
//	 shares a creature type with this creature, draw a card. This ability
//	 triggers only once each turn.""
//
// A layer-6 grant (ADR 0093) to every commander creature the Background's
// controller owns, under anyone's control (Agent of the Iron Throne's shape).
// The trigger is the commander's own: "you" is whoever controls it now and
// "this creature" is the commander as it is now. The spell shares a type by
// game.SharesCreatureType (CR 702.73a: a changeling shares with any
// creature type; a spell with no creature type shares nothing). "Only once
// each turn" limits the TRIGGERING, so it reads the per-turn record of
// triggers put on the stack (b11TriggeredThisTurn, as Exemplar of Light):
// a second matching spell the same turn does nothing, even if the first
// draw was countered.
//
// "Choose a Background" is a deck-construction rule (CR 702.124), the
// deck importer's business.
//
// No simplification. Marvel Super Heroes (#4455).
const (
	msc18FolkHeroGrant = "folk-hero/draw-on-shared-type-spell"
	msc18FolkHeroLabel = "Folk Hero — draw a card"
)

func init() {
	Register(Spec{
		OracleID:     "c5126d6e-969c-4034-98e3-f8653163346f",
		Name:         "Folk Hero",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: msc18FolkHeroGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventCast, msc18CastASpellSharingATypeWithThis, msc18FolkHeroLabel, Do(DrawCards{N: 1})),
			},
			Text: "Whenever you cast a spell that shares a creature type with this creature, draw a card. This ability triggers only once each turn.",
		}},
		Static: []game.StaticAbility{GrantAbilities(commanderCreatureYouOwn, msc18FolkHeroGrant)},
	})
}

// msc18CastASpellSharingATypeWithThis is the granted trigger's condition:
// the commander's controller cast a spell that shares a creature type with
// the commander, and this ability has not triggered yet this turn.
func msc18CastASpellSharingATypeWithThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !game.SharesCreatureType(source, &spell) {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, msc18FolkHeroLabel)
}
