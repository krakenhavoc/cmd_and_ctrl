package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DoublesEntering declares the common "an object entering causes a
// triggered ability of a permanent you control" effect. The game
// supplies the entering object as q.Subject and evaluates this predicate
// against its live characteristics.
func DoublesEntering(filter CardPredicate) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label: "entering",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.HasSubject || !isEntering(q.Event) {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return filter == nil || filter(g, q.DoublerLKI.Controller, subjectCard(g, q))
		},
	}
}

// DoublesDying declares a creature-dying doubler. The subject passed by the
// game is its last-known battlefield object, so an animated land or a creature
// that has already left can still be classified correctly.
func DoublesDying(filter CardPredicate) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label: "dying",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.HasSubject || !isDying(q.Event) {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return filter == nil || filter(g, q.DoublerLKI.Controller, subjectCard(g, q))
		},
	}
}

// DoublesAttacking declares an attacking-object doubler. Tapping caused by
// an attack is deliberately not included; EventAttack is the rules cause.
func DoublesAttacking(filter CardPredicate) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label: "attacking",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.HasSubject || q.Event.Kind != game.EventAttack {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return filter == nil || filter(g, q.DoublerLKI.Controller, subjectCard(g, q))
		},
	}
}

// DoublesAbilitiesOfOptions controls the two exceptional parts of the
// source-based wording. SourceMatch is used by Cloud to relate an Equipment
// source to Cloud itself; ordinary cards should leave it nil and use filter.
// SkipControllerCheck is Cloud's printed exception: its wording names Cloud
// or an attached Equipment without saying "you control".
type DoublesAbilitiesOfOptions struct {
	SourceMatch         func(*game.Game, game.TriggerDoublingQuery) bool
	SkipControllerCheck bool
}

// DoublesAbilitiesOf declares "a triggered ability of a [filter] you
// control". FromSpell is rejected here because these cards refer to
// permanents; the future spell-aware helper will opt into that separately.
func DoublesAbilitiesOf(filter CardPredicate, opts DoublesAbilitiesOfOptions) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label: "ability source",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || q.Ability == nil {
				return false
			}
			if !opts.SkipControllerCheck && q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			if opts.SourceMatch != nil {
				return opts.SourceMatch(g, q)
			}
			return filter == nil || filter(g, q.DoublerLKI.Controller, characteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI))
		},
	}
}

func isEntering(ev game.Event) bool {
	if ev.Kind == game.EventETB || ev.Kind == game.EventTokenCreated {
		return true
	}
	return ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield
}

func subjectCard(g *game.Game, q game.TriggerDoublingQuery) game.Card {
	var fallback game.Card
	if g != nil {
		fallback, _ = g.LookupCardForEffect(q.Subject)
	}
	return characteristicCard(q.Subject, fallback, q.SubjectLKI)
}

func characteristicCard(id uuid.UUID, fallback game.Card, ch game.Characteristic) game.Card {
	c := game.Card{InstanceID: id, Name: ch.Name, Owner: fallback.Owner, Controller: ch.Controller, Power: ch.Power, Toughness: ch.Toughness, Colors: append([]string(nil), ch.Colors...), Keywords: append([]string(nil), ch.Abilities...), ManaCost: fallback.ManaCost}
	// Keep immutable identity and attachment information when the caller's
	// fallback is the live source. The characteristic is authoritative for
	// every layered field, including empty colours and abilities.
	c.OracleID = fallback.OracleID
	c.AttachedTo = fallback.AttachedTo
	parts := append(append([]string(nil), ch.Supertypes...), ch.Types...)
	c.TypeLine = strings.Join(parts, " ")
	if len(ch.Subtypes) > 0 {
		c.TypeLine += " — " + strings.Join(ch.Subtypes, " ")
	}
	return c
}

func isDying(ev game.Event) bool {
	if ev.Kind == game.EventLTB {
		return ev.NewZone == game.ZoneGraveyard
	}
	if ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneBattlefield {
		return false
	}
	return ev.NewZone == game.ZoneGraveyard
}
