package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponentsCreaturesAndNonbasicLandsEnterTapped is "Creatures and nonbasic
// lands your opponents control enter tapped" (Thalia, Heretic Cathar; Thalia
// and The Gitrog Monster): a replacement on the battlefield move of a creature
// or nonbasic land that a player other than the source's controller controls.
func OpponentsCreaturesAndNonbasicLandsEnterTapped(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller == src.Controller {
				return false
			}
			return entering.IsCreature() || B10NonbasicLand()(g, uuid.Nil, entering)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// s82CombatDamageToPlayerOrBattleBy is "a creature you control that matches
// pred dealt combat damage to a player or battle": a combat damage event with
// a positive amount whose source is a creature controlled by controller that
// passes pred and whose recipient is a player or a battle.
func S82CombatDamageToPlayerOrBattleBy(ev game.Event, controller uuid.UUID, g *game.Game, pred func(game.Card) bool) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) == nil {
		target, ok := g.LookupCardForEffect(ev.Target)
		if !ok || !target.IsBattle() {
			return false
		}
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature() && src.Controller == controller && (pred == nil || pred(src))
}

// s82BatchDamageDealers lists, once each and in the order they dealt it, the
// creatures of controller that pass pred and dealt the combat damage to first's
// recipient in first's event batch: the "those Dragons" of a "whenever one or
// more … deal combat damage" ability (CR 603.2c). They are read off the event
// log, so one that has since left the battlefield is still named.
func S82BatchDamageDealers(g *game.Game, first game.Event, controller uuid.UUID, pred func(game.Card) bool) []uuid.UUID {
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		if ev.Batch != first.Batch || ev.Target != first.Target || seen[ev.Source] {
			continue
		}
		if S82CombatDamageToPlayerOrBattleBy(ev, controller, g, pred) {
			seen[ev.Source] = true
			out = append(out, ev.Source)
		}
	}
	return out
}

// s82OtherAttackers lists the attacking creatures other than source, the
// "other attacking creatures" of Ghalta and Mavren, read now at resolution.
func S82OtherAttackers(g *game.Game, source uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source && c.IsCreature() && c.AttackingTarget != uuid.Nil {
			out = append(out, c)
		}
	}
	return out
}
