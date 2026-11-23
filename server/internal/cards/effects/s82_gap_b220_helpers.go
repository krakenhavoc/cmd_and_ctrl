package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b220_helpers.go — helpers for GAP batch 220 (#5948), Commander Legends:
// Battle for Baldur's Gate red creatures.

// GapB220YouCastFromExile is "whenever you cast a spell from exile": a cast by the
// source's controller whose announce record says it came from exile (EventCast
// stamps the origin in OldZone, CR 601.2a).
func GapB220YouCastFromExile(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
}

// GapB220CopyIsAPermanentSpell reports whether the copiable values a copy of a spell is
// being made with describe a permanent spell: not an instant and not a sorcery (CR 304.4,
// CR 307.4). Nalfeshnee's "If it's a permanent spell".
func GapB220CopyIsAPermanentSpell(v *game.PrintedValues) bool {
	line := strings.ToLower(v.TypeLine)
	return !strings.Contains(line, "instant") && !strings.Contains(line, "sorcery")
}

// GapB220OpponentsYouAreAttacking counts the distinct opponents of `you` being attacked
// by `you` right now (CR 506.3d): each defending player of one of your attacking
// creatures, a player whose planeswalker or battle is attacked included, once.
func GapB220OpponentsYouAreAttacking(g *game.Game, you uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.AttackingTarget == uuid.Nil {
			continue
		}
		if p := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil && p != you {
			seen[p] = true
		}
	}
	return len(seen)
}

// GapB220YoureAttackingTwoOrMoreOpponents is Nemesis Phoenix's "only if you're attacking
// two or more opponents" as an activation condition.
func GapB220YoureAttackingTwoOrMoreOpponents(g *game.Game, controller, _ uuid.UUID) bool {
	return GapB220OpponentsYouAreAttacking(g, controller) >= 2
}
