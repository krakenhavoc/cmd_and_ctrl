package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushKeywordCreature puts a 2/2 with printed keywords onto the
// battlefield under `owner`.
func PushKeywordCreature(g *game.Game, owner uuid.UUID, name string, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
}

// creatureTargetableBy reports whether a "target creature" clause
// offers `id` to spells and abilities `by` controls.
func CreatureTargetableBy(g *game.Game, by, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		for _, c := range g.LegalTargetsForEffect(game.TargetSource{Controller: by}, effects.TargetCreature("target creature")).Cards {
			if c == id {
				ok = true
			}
		}
	})
	return ok
}

func PlayerTargetableBy(g *game.Game, by, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		for _, p := range g.LegalTargetsForEffect(game.TargetSource{Controller: by}, effects.TargetPlayer("target player")).Players {
			if p == id {
				ok = true
			}
		}
	})
	return ok
}
