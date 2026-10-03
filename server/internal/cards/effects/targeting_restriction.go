package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// targeting_restriction.go — ADR 0109 §6 (#1885): constructors for
// Spec.TargetingRestrictions, "<these> can't be the targets of spells or
// abilities" about a whole ZONE (CR 601.2c, 608.2b, 101.2). Append-only.
//
// The engine asks the restriction at its two targeting choke points
// (game/targets.go), so a card refused here is never offered by the
// picker or the bot, can't be announced, and makes a spell already aimed
// at it fizzle. A cost, and a "choose" that is not "target" (CR 115.10a),
// is not refused. The predicate runs under g.mu: read-only.

// CardsInGraveyardsCantBeTargeted is Ground Seal's, Dennick's, Silent
// Gravestone's and Underworld Cerberus's "Cards in graveyards can't be
// the targets of spells or abilities." Every card in every graveyard,
// against every spell and ability, its controller's included.
func CardsInGraveyardsCantBeTargeted(label string) game.TargetingRestriction {
	return game.TargetingRestriction{
		Label:   label,
		Zones:   []game.ZoneKind{game.ZoneGraveyard},
		Forbids: func(game.TargetingQuery) bool { return true },
	}
}

// OpponentsCantTarget is "<these> can't be the targets of spells or
// abilities your opponents control": a candidate in one of `zones` that
// `match` accepts, against a spell or ability controlled by anyone but
// the source's controller. Tomik's "Lands on the battlefield and land
// cards in graveyards" is Land() over the battlefield and the
// graveyards. `match` reads the candidate as it stands in its zone, so a
// permanent is judged by its effective types and a graveyard card by its
// printed ones.
func OpponentsCantTarget(label string, match CardPredicate, zones ...game.ZoneKind) game.TargetingRestriction {
	return game.TargetingRestriction{
		Label: label,
		Zones: zones,
		Forbids: func(q game.TargetingQuery) bool {
			if q.Controller == q.Source.Controller {
				return false
			}
			return match == nil || match(q.Game, q.Controller, q.Card)
		},
	}
}
