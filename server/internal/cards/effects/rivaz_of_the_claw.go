package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rivaz of the Claw — Legendary Creature — Lizard Warlock {1}{B}{R}, 3/3
// (Dominaria United; deck request #4724):
//
//	"Menace
//	 {T}: Add two mana in any combination of colors. Spend this mana
//	 only to cast Dragon creature spells.
//	 Once during each of your turns, you may cast a Dragon creature
//	 spell from your graveyard.
//	 Whenever you cast a Dragon creature spell from your graveyard, it
//	 gains "When this creature dies, exile it.""
//
// Menace is the printed keyword. The mana is Gwenna's two independent
// any-colour slots (AnyCombinationOfColors) carrying Carnelian Orb's
// three spend tags, which AND: it pays only for casting a spell that is
// a creature AND a Dragon (CR 106.6), so not a Dragon's activated
// ability, not a non-Dragon creature and not a noncreature Dragon spell.
// As a creature's {T} ability it waits out summoning sickness.
//
// The graveyard cast is Gisa and Geralf's standing permission (ADR 0066)
// with the creature type swapped: one cast per turn per Rivaz OBJECT
// (CastPermission.PerType holding "creature"), closed off your turn
// (TimingYourTurnOnly), with the Dragon's own timing and costs in force,
// so a second Rivaz opens a second cast as the rulings say.
//
// The last ability is a cast trigger told the spell's origin by the
// stack item's CastFromZone, the Devoted Grafkeeper reading: a Dragon
// creature spell cast from your own graveyard by ANY permission (Rivaz's,
// a flashback-granting card, another Rivaz) triggers it; one cast from
// hand, exile or an opponent's graveyard does not. As it resolves, the
// spell gains the bundle below. That is an ability-adding effect on a
// permanent spell, which the permanent it becomes keeps (CR 400.7a): one
// layer-6 scoped effect with GrantAbilitiesMod, held for as long as the
// spell is on the stack and then the permanent is on the battlefield as
// that object (ForThePermanentThisSpellBecomesDuration, the Strago and
// Relm rider's shape), whoever controls it, and no longer tied to Rivaz.
// A spell countered or gone before the trigger resolves gains nothing.
//
// The granted "When this creature dies, exile it." is the creature's own
// trigger (ADR 0093): it fires from last-known information, is
// controlled by the creature's controller, and exiles the card only
// while it is still the object that died in a graveyard (CR 400.7), so a
// commander its owner moved to the command zone, or a card reanimated in
// response, is left alone.
//
// No simplification.
const rivazOfTheClawDiesExileGrant = "rivaz-of-the-claw/dies-exile-it"

func init() {
	Register(Spec{
		OracleID:        "ecff46b9-fcc9-4895-8b6f-c697fd15d01a",
		Name:            "Rivaz of the Claw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     AnyCombinationOfColors(2),
			Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype("Dragon")},
			Label:        "Add two mana in any combination of colors. Spend this mana only to cast Dragon creature spells",
		}},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{CreatureOnly: true, CreatureType: "Dragon"},
			PerType:  []string{game.PermissionTypeCreature},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Once during each of your turns, cast a Dragon creature spell from your graveyard (Rivaz of the Claw)",
		}},
		Grants: []AbilityGrant{{
			Key: rivazOfTheClawDiesExileGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Rivaz of the Claw — this creature died, exile it", rivazExileTheCreatureThatDied),
			},
			Text: "When this creature dies, exile it.",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, rivazYouCastADragonCreatureFromYourGraveyard,
				"Rivaz of the Claw — it gains \"When this creature dies, exile it.\"",
				rivazThatSpellGainsDiesExile),
		},
	})
}

// rivazYouCastADragonCreatureFromYourGraveyard is "whenever you cast a
// Dragon creature spell from your graveyard": the Grafkeeper's graveyard
// cast, narrowed to a spell that is a creature and a Dragon as it sits on
// the stack.
func rivazYouCastADragonCreatureFromYourGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !youCastASpellFromYourGraveyard(ev, source, lki, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	// "Dragon" is a word of the source's text (CR 612.2).
	return ok && spell.IsCreature() && spell.HasSubtype(source.TextSubtype("Dragon"))
}

// rivazThatSpellGainsDiesExile gives the spell the trigger named, while
// it is still a spell on the stack, the dies-exile bundle for the
// permanent it becomes (CR 400.7a).
func rivazThatSpellGainsDiesExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell := item.Trigger.Event.CardID
	if spell == uuid.Nil || g.StackItemForEffect(spell) == nil {
		return nil
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]game.AffectedObject{{ID: spell}},
		[]game.Mod{game.GrantAbilitiesMod(rivazOfTheClawDiesExileGrant)},
		game.ForThePermanentThisSpellBecomesDuration(spell),
		"Rivaz of the Claw — it gains \"When this creature dies, exile it.\"")
	return nil
}

// rivazExileTheCreatureThatDied is the granted trigger's body: the card
// that died, while it is still that object in a graveyard, is exiled.
func rivazExileTheCreatureThatDied(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	return ExileTarget{Target: dead}.Apply(ctx)
}
