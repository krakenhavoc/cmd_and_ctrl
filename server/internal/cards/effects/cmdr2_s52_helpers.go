package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S52CombatDamageToPlayerDoubler is Felix Five-Boots' "If a creature you
// control dealing combat damage to a player causes a triggered ability of a
// permanent you control to trigger, that ability triggers an additional time"
// (CR 603.2d). The cause is a combat EventDealDamage whose source is a
// creature the doubler's controller controls (read from its last-known
// characteristics) and whose target is a player; the ability belongs to a
// permanent they control.
func cmdr2S52CombatDamageToPlayerDoubler(name string) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    name,
		RowLabel: "Abilities of your permanents that trigger when a creature you control deals combat damage to a player trigger an additional time.",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.SourceIsPermanent || q.Ability == nil || !q.HasSubject {
				return false
			}
			ev := q.Event
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return q.SubjectLKI.Controller == q.DoublerLKI.Controller && slices.Contains(q.SubjectLKI.Types, "Creature")
		},
	}
}

// cmdr2S52OpponentsBelowHalfStarting counts the source controller's living
// opponents whose life total is less than half their starting life total
// (Anya, Merciless Angel). The table's starting total is read through
// startingLifeOf; the comparison is life*2 < starting so an odd starting
// total has no rounding.
func cmdr2S52OpponentsBelowHalfStarting(g *game.Game, source *game.Card) int {
	if g == nil || source == nil {
		return 0
	}
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.ID == source.Controller || p.Eliminated {
			continue
		}
		if p.Life*2 < startingLifeOf(g) {
			n++
		}
	}
	return n
}
