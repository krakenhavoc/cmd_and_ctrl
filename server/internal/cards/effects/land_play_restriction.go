package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_play_restriction.go — ADR 0109 §4 (#1895): constructors for
// Spec.LandPlayRestrictions ("players can't play lands", CR 101.2) and
// the "this turn" effect that writes one onto a player (Turf Wound).
//
// Playing a land is a special action, not a cast (CR 305.1, CR 116.2a),
// so a "can't cast" restriction never reaches one. These are the twins
// of cast_restriction.go's constructors, one per printed SHAPE, for the
// same reason: the shape carries the clause the player is shown, and
// Register refuses a restriction with no label or no predicate.
//
// The predicate runs under g.mu inside the play path. Read-only:
// *ForEffect accessors and plain field reads.

// PlayersCantPlayLands is Territorial Dispute's and Worms of the Earth's
// "Players can't play lands." No player is named, so it binds every
// player, the permanent's controller included (CR 101.2: "can't" beats
// the land drop, and an extra drop does not lift it).
func PlayersCantPlayLands(label string) game.LandPlayRestriction {
	return game.LandPlayRestriction{
		Label:   label,
		Forbids: func(game.LandPlayQuery) bool { return true },
	}
}

// YouCantPlayLands is Aggressive Mining's "You can't play lands." "You"
// is the permanent's own CONTROLLER, not the player playing the land, so
// an Aggressive Mining an opponent has stolen restricts THEM.
func YouCantPlayLands(label string) game.LandPlayRestriction {
	return game.LandPlayRestriction{
		Label: label,
		Forbids: func(q game.LandPlayQuery) bool {
			return q.Source.Controller == q.Player
		},
	}
}

// OpponentsCantPlayLandsFrom is Tomik's "Your opponents can't play land
// cards from graveyards": every player except the source's controller,
// for a land played out of one of `zones`.
func OpponentsCantPlayLandsFrom(label string, zones ...game.ZoneKind) game.LandPlayRestriction {
	banned := make(map[game.ZoneKind]bool, len(zones))
	for _, z := range zones {
		banned[z] = true
	}
	return game.LandPlayRestriction{
		Label: label,
		Forbids: func(q game.LandPlayQuery) bool {
			return q.Source.Controller != q.Player && banned[q.FromZone]
		},
	}
}

// CantPlayLandsNamed is City in a Bottle's "Players can't … play lands
// with a name originally printed in the Arabian Nights expansion": every
// player, for a land whose name `isName` accepts (game.IsArabianNightsName,
// CR 206.3a's list). The land is judged by the name of the face being
// played.
func CantPlayLandsNamed(label string, isName func(name string) bool) game.LandPlayRestriction {
	return game.LandPlayRestriction{
		Label: label,
		Forbids: func(q game.LandPlayQuery) bool {
			return isName(q.Card.Name)
		},
	}
}

// OpponentsWithMoreLandsCantPlayLands is Ward of Bones's "Each opponent
// who controls more lands than you can't play lands": every player but
// the source's controller, who controls more lands than that controller
// does right now.
func OpponentsWithMoreLandsCantPlayLands(label string) game.LandPlayRestriction {
	return game.LandPlayRestriction{
		Label: label,
		Forbids: func(q game.LandPlayQuery) bool {
			return q.Source.Controller != q.Player &&
				controlsMoreThan(q.Game, q.Player, q.Source.Controller, QueryType("land"))
		},
	}
}

// OpponentsWithMoreCantCast is Ward of Bones's "Each opponent who
// controls more creatures than you can't cast creature spells" (and its
// artifact and enchantment twins): every player but the source's
// controller, who controls more permanents matching `q` than that
// controller does, for a spell `match` accepts.
func OpponentsWithMoreCantCast(label string, q game.PermanentQuery, match CardPredicate) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(cq game.CastQuery) bool {
			return cq.Source.Controller != cq.Controller &&
				controlsMoreThan(cq.Game, cq.Controller, cq.Source.Controller, q) &&
				matchCastCard(cq.Game, match, cq.Controller, cq.Card)
		},
	}
}

// controlsMoreThan reports whether `player` controls strictly more
// permanents matching q than `other` does.
func controlsMoreThan(g *game.Game, player, other uuid.UUID, q game.PermanentQuery) bool {
	return g.CountControlledMatchingForEffect(player, q) > g.CountControlledMatchingForEffect(other, q)
}

// permanentWasCastThisTurn is "this creature was cast this turn" for a
// permanent on the battlefield (Rock Jockey): the engine keeps no
// per-card cast fact, so it is read off the turn's event log, which is
// already bounded to the current turn. The permanent was cast this turn
// if an EventCast names it and it has not left the battlefield since: a
// permanent that was bounced and flickered back is a new object (CR
// 400.7) that was not cast.
func permanentWasCastThisTurn(g *game.Game, id uuid.UUID) bool {
	cast := false
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID != id {
			continue
		}
		switch ev.Kind {
		case game.EventCast:
			cast = true
		case game.EventLTB:
			cast = false
		}
	}
	return cast
}

// CantPlayLandsThisTurn is "Target player can't play lands this turn"
// (Turf Wound, Solfatara, Pardic Miner, Moonhold): each still-legal
// player target gets a ModCantPlayLands record, swept at cleanup (CR
// 514.2). A target that has left the game is skipped (CR 608.2b).
//
// It is a stored record and not a static because the spell that wrote it
// is gone the moment it resolves; the record is what keeps the ban alive.
type CantPlayLandsThisTurn struct{}

func (CantPlayLandsThisTurn) Apply(ctx *Context) error {
	label := "can't play lands this turn"
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		ctx.Game.CantPlayLandsThisTurnForEffect(ctx.Source(), t.ID, label)
	}
	return nil
}
