package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sevinne, the Chronoclasm — Legendary Creature — Human Wizard {2}{U}{R}{W},
// 2/2 (Commander 2019):
//
//	"Prevent all damage that would be dealt to Sevinne.
//	 Whenever you cast your first instant or sorcery spell from your
//	 graveyard each turn, copy that spell. You may choose new targets for
//	 the copy."
//
// The prevention is Glittering Lion's standing self prevention. The trigger
// reads the caster's per-turn tally of instants and sorceries cast from the
// graveyard (CastTally.InstantSorceryFromGraveyard, bumped as each spell
// becomes cast, CR 601.2i), so it fires for the first only; the spell must
// come from the caster's own graveyard (it is one of their cards). The copy
// is created, not cast (CR 707.10), so it triggers nothing, and it is made
// from last-known information if the original has already left the stack
// (CR 608.2h), with the CR 707.10c retarget choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "869c9bc4-2b21-40c3-8b1c-c83269c856d0",
		Name:         "Sevinne, the Chronoclasm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				Label: "Sevinne, the Chronoclasm — prevent all damage that would be dealt to it",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, sevinneFirstFromYourGraveyard,
				"Sevinne, the Chronoclasm — copy that spell; you may choose new targets for the copy",
				effects.CopyTheTriggeringSpellWithNewTargets),
		},
	})
}

// sevinneFirstFromYourGraveyard is "you cast your first instant or sorcery
// spell from your graveyard each turn".
func sevinneFirstFromYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || ev.OldZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Owner != source.Controller || !(c.IsInstant() || c.IsSorcery()) {
		return false
	}
	return g.SpellsCastThisTurn[source.Controller].InstantSorceryFromGraveyard == 1
}
