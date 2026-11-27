package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB609PlayerCastsMayPayGainLife is the Eighth Edition life-gain artifacts'
// trigger (Crystal Rod, Iron Star, Throne of Bone): "Whenever a player casts a
// <colour> spell, you may pay {1}. If you do, you gain 1 life." Any player's
// cast counts, the source's controller included; the colour is a word of the
// source's text, read through it (CR 612.2). The {1} is asked as the trigger
// resolves and the life is gained only once it has been paid (CR 603.5).
func OldB609PlayerCastsMayPayGainLife(name, color string, paid game.PayUnlessElse) game.TriggeredAbility {
	return On(game.EventCast, TmcB8APlayerCast(OfColor(color)),
		name+" — you may pay {1}: gain 1 life",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{Chooser: item.Controller, Cost: "{1}", Question: name + " — pay {1} to gain 1 life?", Then: paid}.Apply(NewContext(g, item))
		})
}

// OldB609LifeThreeOrLessSix is Avatar of Hope's reduction amount: {6} when
// the caster has 3 or less life, none otherwise.
func OldB609LifeThreeOrLessSix(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	if p := q.Game.PlayerByIDForEffect(q.Controller); p == nil || p.Life > 3 {
		return 0
	}
	return 6
}
