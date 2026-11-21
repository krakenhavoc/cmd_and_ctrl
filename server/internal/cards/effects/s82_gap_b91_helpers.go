package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB91FirstHistoricSpellThisTurn is "the first historic spell you cast each
// turn" as a spell filter (CR 700.6): the spell is historic and the caster
// cast no other historic spell this turn. The turn's cast events are read, so
// a historic spell that was countered or has resolved still counts, and the
// spell being priced is not counted against itself.
func gapB91FirstHistoricSpellThisTurn(g *game.Game, caster uuid.UUID, spell game.Card) bool {
	if !b09IsHistoric(spell) {
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != caster || ev.CardID == spell.InstanceID {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && b09IsHistoric(c) {
			return false
		}
	}
	return true
}

// gapB91ExileTheChosenTarget is "exile target <thing>" as a whole trigger body
// (Atraxi Warden's entry, Archon of Justice's death): the first target still
// legal as the ability resolves is exiled (CR 608.2b).
func gapB91ExileTheChosenTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ExileTarget{Target: id}.Apply(ctx)
}
