package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82MidB17SourceIsSpell reports whether the object a damage event names
// as its source is a SPELL right now: a card in the stack zone. A spell
// stays in the stack zone while its effect runs, and is routed to its
// graveyard only afterwards (CR 608.2n), so "a spell deals damage" reads
// the zone it is dealing it from. A permanent's ability, including one a
// creature that has just died deals through, names a source that is not
// in the stack zone, so it is never taken for a spell.
func s82MidB17SourceIsSpell(g *game.Game, source uuid.UUID) bool {
	if source == uuid.Nil {
		return false
	}
	z := g.FindCardZoneForEffect(source)
	return z != nil && z.Kind == game.ZoneStack
}

// s82MidB17DamageSourceSpell is the spell a damage event comes from: its
// card as it is now, and whether the event's source is a spell at all.
func s82MidB17DamageSourceSpell(g *game.Game, source uuid.UUID) (game.Card, bool) {
	if !s82MidB17SourceIsSpell(g, source) {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(source)
}

// s82MidB17SpellYouControlDealtDamage is "a spell you control deals
// damage": the damage event's source is a spell its controller is the
// source permanent's controller (Flame Channeler, Embodiment of Flame).
func s82MidB17SpellYouControlDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || source == nil {
		return false
	}
	spell, ok := s82MidB17DamageSourceSpell(g, ev.Source)
	return ok && spell.Controller == source.Controller
}

// s82MidB17WheneverASpellYouControlDealsDamage is the trigger both faces
// of Flame Channeler share. A spell that deals damage to several things
// at once deals it in one occurrence (CR 120.4d), so the ability is
// once per damage event of one spell: OncePerBatch keyed by the spell.
func s82MidB17WheneverASpellYouControlDealsDamage(label string, effect Effect) game.TriggeredAbility {
	t := OncePerBatch(On(game.EventDealDamage, s82MidB17SpellYouControlDealtDamage, label, effect))
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Source.String() }
	return t
}
