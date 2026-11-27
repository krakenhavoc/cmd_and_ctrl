package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Echoes — Enchantment — Aura Curse {4}{U} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Enchant player
//	 Whenever enchanted player casts an instant or sorcery spell, each other
//	 player may copy that spell and may choose new targets for the copy they
//	 control."
//
// Wandering Archaic's copy question, asked once of every player but the
// caster in turn order starting from the active player (CR 101.4): the spell
// is held for last-known information (it may resolve before the table has
// answered, #4276) and each "yes" is a separate copy that player controls,
// with the CR 707.10c prompt to choose new targets. The enchanted player's
// own copy is not offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c99bac4b-7eec-4fdf-b344-0cf9639ff073",
		Name:         "Curse of Echoes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer || ev.Actor != source.AttachedTo.ID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			}, "Curse of Echoes — each other player may copy that spell", effects.S82OldB440EachOtherPlayerMayCopy(curseOfEchoesCopy)),
		},
	})
}

// curseOfEchoesCopy is one player's answer to "copy that spell?": the copy is
// theirs. An on-disk key: never renamed, never reused.
var curseOfEchoesCopy = effects.MayAnswer("confirm/curse-of-echoes-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	return effects.CopySpell{
		StackID:          r.Carry[0],
		Controller:       r.Chooser,
		FromLastKnown:    true,
		ChooseNewTargets: true,
	}.Apply(ctx)
})
