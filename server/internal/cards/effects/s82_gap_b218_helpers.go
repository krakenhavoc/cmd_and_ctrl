package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b218_helpers.go — helpers for GAP batch 218 (#5948), Commander
// Legends: Battle for Baldur's Gate legends and red creatures.

// GapB218CastFromExileOrLandEntersFromExile is Faldorn's condition: "you cast a
// spell from exile or a land you control enters from exile". A cast reads the
// origin the announce record stamped in OldZone (CR 601.2a); a land reads a zone
// move from exile onto the battlefield (CR 400.7), however it got there, with
// the land under the source's controller's control as it entered.
func GapB218CastFromExileOrLandEntersFromExile(ev game.Event, source *game.Card, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
	case game.EventZoneMove:
		if ev.OldZone != game.ZoneExile || ev.NewZone != game.ZoneBattlefield {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsLand() && c.Controller == source.Controller
	}
	return false
}

// GapB218SpellWithCascade is "a spell with cascade": a card whose own triggered
// abilities include the cascade keyword's (Aurora Phoenix, The First Doctor).
func GapB218SpellWithCascade(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	for _, t := range game.TriggersForCard(c) {
		if t.Keyword == KeywordCascade {
			return true
		}
	}
	return false
}

// GapB218IsEnchanted reports whether an Aura is attached to the permanent
// (an Equipment does not enchant), read through effective subtypes.
func GapB218IsEnchanted(g *game.Game, c game.Card) bool { return b15IsEnchanted(g, c) }

// GapB218PlayersBeingAttacked counts the distinct players being attacked right
// now (CR 506.3d): each defending player of every attacking creature, a player
// whose planeswalker or battle is attacked included, once however many
// creatures attack them. Amber Gristle O'Maul's "a card for each player being
// attacked".
func GapB218PlayersBeingAttacked(g *game.Game) int {
	seen := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if p := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil {
			seen[p] = true
		}
	}
	return len(seen)
}

// GapB218HasAManaAbility is "a creature with a mana ability" (CR 605.1): an
// activated or triggered mana ability it has right now, printed or granted by an
// effect (ManaAbilitiesForCard reads the layered view).
func GapB218HasAManaAbility(c game.Card) bool {
	return len(game.ManaAbilitiesForCard(c)) > 0 || len(game.ManaTriggersForCard(c)) > 0
}
