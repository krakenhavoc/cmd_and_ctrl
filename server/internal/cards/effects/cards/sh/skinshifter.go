package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skinshifter — Creature — Human Shaman {1}{G}, 1/1 (Magic 2012, S82 older
// sets batch 451, #5573):
//
//	"{G}: Choose one. Activate only once each turn.
//	 • Until end of turn, this creature becomes a Rhino with base power and
//	   toughness 4/4 and gains trample.
//	 • Until end of turn, this creature becomes a Bird with base power and
//	   toughness 2/2 and gains flying.
//	 • Until end of turn, this creature becomes a Plant with base power and
//	   toughness 0/8."
//
// Each mode is one record with three modifications at one timestamp: the
// creature types (layer 4, replacing the old ones), the base power and
// toughness (layer 7b) and the keyword (layer 6). +1/+1 counters and other
// boosts still apply on top of the new base size.
//
// No simplification.
func init() {
	const label = "{G}: Choose one. Activate only once each turn."
	becomes := func(name, kind string, power, toughness int, keywords ...string) game.ModeOption {
		return effects.ModeDoing("Until end of turn, this creature becomes a "+kind+" with base power and toughness "+
			effects.S82OldB451PTWords(power, toughness)+effects.S82OldB451GainsWords(keywords)+".", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.SourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				mods := []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes(kind)...)}
				mods = append(mods, game.SetBasePTMods(power, toughness)...)
				if len(keywords) > 0 {
					mods = append(mods, game.AddKeywordsMod(keywords...))
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     mods,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    name + " — becomes a " + kind,
				}.Apply(ctx)
			})
	}
	effects.Register(effects.Spec{
		OracleID:     "cfa3f2a4-2724-4bc1-bdaf-244051db95ed",
		Name:         "Skinshifter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Modes: effects.ChooseOne(
				becomes("Skinshifter", "Rhino", 4, 4, "trample"),
				becomes("Skinshifter", "Bird", 2, 2, "flying"),
				becomes("Skinshifter", "Plant", 0, 8),
			),
		}},
	})
}
