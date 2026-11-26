package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Processor Assault — Sorcery {1}{R} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 As an additional cost to cast this spell, put a card an opponent owns from
//	 exile into that player's graveyard.
//	 Processor Assault deals 5 damage to target creature."
//
// The additional cost is Void Attendant's processing component (the caster
// names the card on cast_spell as exile_ids; with none the spell can't be
// cast, CR 118.3), paid as the spell is cast (CR 601.2h) and charged under
// every alternative cost. The damage is Lightning Bolt's, at 5.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e773258a-e59b-4e5f-b4e1-4f95cb11a39f",
		Name:            "Processor Assault",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		AdditionalCost:  effects.OldB308ProcessACardCost(),
		Targets:         effects.TargetCreature("target creature"),
		Purpose:         effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:       effects.DamageToFirstTargetStep(5),
	})
}
