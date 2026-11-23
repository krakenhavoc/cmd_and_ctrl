package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peace Talks — Sorcery {1}{W} (Visions, Multiverse Gift Box; #5718):
//
//	"This turn and next turn, creatures can't attack, and players and
//	 permanents can't be the targets of spells or activated abilities."
//
// Two records with one duration, DurationThisTurnAndNextTurn: the turn
// it resolves in and the turn after it, whoever takes that turn, ending
// in that turn's cleanup step (the 2004-10-04 ruling: "exactly two turns,
// not a full round of turns"; an extra turn right after this one is the
// next turn).
//
//   - "Creatures can't attack" is can't-attack on every creature, read
//     live (CR 611.2c: it changes no characteristic), so a creature that
//     arrives or becomes a creature later can't attack either. The combat
//     phase still happens (the 2006-02-01 ruling).
//   - "Players and permanents can't be the targets of spells or activated
//     abilities" is the table-wide targeting bar
//     (Game.PlayersAndPermanentsCantBeTargetedForEffect): every player and
//     every permanent, read live, against a spell (an Aura spell
//     included, CR 303.4a) or an activated ability (equip included); a
//     triggered ability still targets. It is asked by the engine's one
//     targeting walk, so the picker, the bot, the announce gate (CR
//     601.2c) and the resolution re-check (CR 608.2b) agree. Spells on the
//     stack and cards in other zones are neither, so a counterspell still
//     targets a spell.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0f87e25-6604-4723-9467-1e4fedc1157a",
		Name:         "Peace Talks",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			d := effects.DurationThisTurnAndNextTurn(ctx)
			ctx.Game.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeCreatures, item.Controller,
				[]game.Mod{game.AddRestrictionsMod(game.CantAttack)}, d,
				"Peace Talks — creatures can't attack this turn and next turn")
			ctx.Game.PlayersAndPermanentsCantBeTargetedForEffect(item.SourceCardID, item.Controller, d,
				"Peace Talks — players and permanents can't be the targets of spells or activated abilities this turn and next turn")
			return nil
		},
	})
}
