package game

import (
	"strconv"

	"github.com/google/uuid"
)

// as_cast_opponent.go — #5860: a target clause that exists only for an
// opponent who met a condition AS THE SPELL WAS CAST.
//
//	"Corrupted — For each opponent who has three or more poison
//	 counters as you cast this spell, put up to one target creature
//	 card from that player's graveyard onto the battlefield under your
//	 control." (Geth's Summons)
//
// The clauses are opponent_clauses.go's "one per possible opponent, each
// bound to the n-th seat after the chooser's, each up to one", with one
// more fact: whether that opponent had the poison as the spell was cast.
// Targets are announced during the cast (CR 601.2c), so the announce
// walk judges the poison live, which IS "as you cast". The answer is
// then fixed: CastSpell stamps it on the item next to the cast-time
// values (CR 601.2i, cast_time_values.go), and every later walk of the
// item -- the CR 608.2b re-check, a CR 115.7 retarget, a CR 707.10
// copy's new targets -- reads the stamp, never the live count. An
// opponent who loses their poison in response still has their target;
// one who gains it in response gets none.
//
// DATA on the clause (TargetSpec.AsCastOpponentSlot / AsCastMinPoison),
// judged in the one shared walk (sourceAdmits), so the legal set, the
// announce check, the re-check, the client's picker and the bot's moves
// are one answer. The stamp rides StackItem.CastValues, which Clone,
// spell copies and the snapshot already carry, so a restore mid-stack
// keeps it (additive: new keys in an existing map).

// asCastPoisonKey is the CastValues key recording whether `player` had
// at least `min` poison counters as the spell was cast: 1 yes, 0 no.
// Engine-written, never declared by a card.
func asCastPoisonKey(min int, player uuid.UUID) string {
	return "as-cast-opponent/poison-" + strconv.Itoa(min) + "/" + player.String()
}

// asCastOpponentAdmits is the clause's candidate gate: true without a
// slot. With one, the candidate must be owned by the live opponent in
// that slot -- a card in a graveyard is in its owner's (CR 404.1) -- and
// that opponent must have had AsCastMinPoison or more poison counters as
// the spell was cast: read from the item's stamp when the walk carries
// one, live while the spell is still being cast. An absent seat admits
// nothing.
func (s *TargetSpec) asCastOpponentAdmits(g *Game, src TargetSource, c Card) bool {
	if s.AsCastOpponentSlot < 1 {
		return true
	}
	p := g.OpponentSeatAfter(src.Controller, s.AsCastOpponentSlot)
	if p == nil || c.Owner != p.ID {
		return false
	}
	if v, ok := src.castValues[asCastPoisonKey(s.AsCastMinPoison, p.ID)]; ok {
		return v == 1
	}
	return p.Poison >= s.AsCastMinPoison
}

// stampAsCastOpponentsLocked records, for every opponent of the caster,
// whether they met each as-cast poison threshold the spell's clauses
// name. Called beside stampCastTimeValuesLocked, with every target
// announced and every cost paid; a spell with no such clause records
// nothing and serialises exactly as before.
//
// Caller must hold g.mu.
func (g *Game) stampAsCastOpponentsLocked(card Card, item *StackItem) {
	if item == nil {
		return
	}
	key := CatalogKey(card)
	if key == "" {
		return
	}
	var mins []int
	note := func(spec *TargetSpec) {
		for i := 0; i < spec.ClauseCount(); i++ {
			if cl := spec.Clause(i); cl.AsCastOpponentSlot > 0 && cl.AsCastMinPoison > 0 {
				mins = append(mins, cl.AsCastMinPoison)
			}
		}
	}
	spec, ms := castClauseSources(key)
	if spec != nil {
		note(spec)
	}
	if ms != nil {
		for _, m := range item.Modes {
			if m >= 0 && m < len(ms.Options) && ms.Options[m].Targets != nil {
				note(ms.Options[m].Targets)
			}
		}
	}
	if len(mins) == 0 {
		return
	}
	if item.CastValues == nil {
		item.CastValues = map[string]int{}
	}
	for _, p := range g.Seats {
		if p == nil || p.ID == item.Controller {
			continue
		}
		for _, min := range mins {
			v := 0
			if p.Poison >= min {
				v = 1
			}
			item.CastValues[asCastPoisonKey(min, p.ID)] = v
		}
	}
}
