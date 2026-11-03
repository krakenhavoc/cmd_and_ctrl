package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hofri Ghostforge — Legendary Creature — Dwarf Cleric {3}{R}{W}, 4/5:
//
//	"Spirits you control get +1/+1 and have trample and haste.
//	 Whenever another nontoken creature you control dies, exile it. If
//	 you do, create a token that's a copy of that creature, except it's a
//	 Spirit in addition to its other types and it has "When this token
//	 leaves the battlefield, return the exiled card to its owner's
//	 graveyard.""
//
// The statics are a tribal anthem and two keyword grants over the Spirits
// Hofri's controller controls, read through the layered view.
//
// The trigger exiles the card only while it is still the object that died
// (CR 400.7); a card that has already left its graveyard is not exiled and
// no token is made (the card's rulings). The copy is the creature's
// copiable values as it last existed on the battlefield (CR 707.2,
// 608.2h), so a dead Clone makes what it was copying, and counters, Auras
// and other effects are not copied. "Spirit in addition to its other
// types" and the leaves trigger are part of the copy (CR 707.9a, 707.9b),
// so a copy of the token is a Spirit with that ability too.
//
// The leaves trigger is linked to Hofri's ability (CR 607.2a): it returns
// only the card exiled by the resolution that created that token, and only
// while that card is still in exile as the object it became. It finds the
// pair in the event log, the window from that resolution's EventResolve to
// the next one. Several tokens made by one resolution (a token doubler)
// all point at the same card; a copy of the token made any other way
// points at none and returns nothing, as the rulings say.
//
// No simplification.
func init() {
	spirits := TribeFilter{Tribes: []string{"Spirit"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "663d451b-f848-41cc-b541-0c6fe37d3d54",
		Name:         "Hofri Ghostforge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(spirits, 1, 1),
			TribalKeywordGrant(spirits, "trample"),
			TribalKeywordGrant(spirits, "haste"),
		},
		Grants: []AbilityGrant{{
			Key: hofriGhostforgeReturnGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisLeaves("Hofri Ghostforge — when this token leaves the battlefield, return the exiled card to its owner's graveyard",
					hofriGhostforgeReturnExiled),
			},
			Text: "When this token leaves the battlefield, return the exiled card to its owner's graveyard.",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, hofriGhostforgeAnotherNontokenCreatureDied, hofriGhostforgeLabel, hofriGhostforgeExile),
		},
	})
}

const (
	hofriGhostforgeLabel       = "Hofri Ghostforge — exile it; if you do, create a token that's a copy of that creature"
	hofriGhostforgeReturnGrant = "hofri-ghostforge/return-exiled"
)

// hofriGhostforgeCopyTail is "if you do, create a token that's a copy of
// that creature, except …", run once the exile has settled.
var hofriGhostforgeCopyTail = CardTail("hofri-ghostforge/copy", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(obj.Ref())
	if !ok {
		return nil
	}
	hofriGhostforgeException(&tmpl)
	return createTokenCopies(ctx, ctx.Controller(), tmpl, 1)
})

// hofriGhostforgeException is "except it's a Spirit in addition to its
// other types and it has '…'".
func hofriGhostforgeException(t *game.Card) {
	v := game.PrintedValues{TypeLine: t.TypeLine, GrantedAbilities: t.GrantedAbilities}
	v.AddSubtype("Spirit")
	v.GrantAbility(hofriGhostforgeReturnGrant)
	t.TypeLine = v.TypeLine
	t.GrantedAbilities = v.GrantedAbilities
}

// hofriGhostforgeAnotherNontokenCreatureDied is "another nontoken
// creature you control dies", read as the creature last existed.
func hofriGhostforgeAnotherNontokenCreatureDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID || !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// hofriGhostforgeExile exiles the card that died, if it is still that
// object in its graveyard, and copies it once it is in exile.
func hofriGhostforgeExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	return ExileTarget{Target: dead, ThenTail: ctx.ItemTail(hofriGhostforgeCopyTail)}.Apply(ctx)
}

// hofriGhostforgeReturnExiled is the token's leaves trigger: the card
// Hofri exiled in the resolution that created this token goes to its
// owner's graveyard, if it is still in exile and has not moved since.
func hofriGhostforgeReturnExiled(g *game.Game, item *game.StackItem) error {
	card, ok := hofriGhostforgeLinkedCard(g, item.SourceCardID)
	if !ok {
		return nil
	}
	return g.PutIntoGraveyardForEffect(card)
}

// hofriGhostforgeLinkedCard walks the log for the Hofri resolution in
// which `token` entered, and answers the card that resolution exiled,
// provided nothing has moved it since.
func hofriGhostforgeLinkedCard(g *game.Game, token uuid.UUID) (uuid.UUID, bool) {
	open := false
	var exiled, linked uuid.UUID
	found := false
	for _, ev := range g.Events {
		switch ev.Kind {
		case game.EventResolve:
			open = ev.Label == hofriGhostforgeLabel
			exiled = uuid.Nil
			continue
		case game.EventZoneMove:
			if found {
				if ev.CardID == linked {
					// It left exile after the link was made: a new
					// object, no longer "the exiled card" (CR 400.7).
					return uuid.Nil, false
				}
				continue
			}
			if open && ev.NewZone == game.ZoneExile && exiled == uuid.Nil {
				exiled = ev.CardID
			}
		case game.EventETB:
			if !found && open && ev.CardID == token && exiled != uuid.Nil {
				linked, found = exiled, true
			}
		}
	}
	if !found {
		return uuid.Nil, false
	}
	z := g.FindCardZoneForEffect(linked)
	return linked, z != nil && z.Kind == game.ZoneExile
}
