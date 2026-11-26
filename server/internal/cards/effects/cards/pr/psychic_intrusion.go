package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Intrusion — Sorcery {3}{U}{B} (Theros Beyond Death, S82 old sets):
//
//	"Target opponent reveals their hand. You choose a nonland card from that
//	 player's graveyard or hand and exile it. You may cast that card for as
//	 long as it remains exiled, and you may spend mana as though it were mana
//	 of any color to cast that spell."
//
// Covetous Urge's text: the revealed-hand pick with the graveyard, then a
// cast-only permission for the chooser, for as long as the card remains
// exiled, with mana of any color. The helper is shared; the continuation is
// keyed under this card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdb10291-b5e0-4da8-a103-8e7ac7f8177b",
		Name:         "Psychic Intrusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB397RevealNonlandPick(ctx, psychicIntrusionMayCast)
		},
	})
}

// psychicIntrusionMayCast is the "you may cast that card" continuation. An
// on-disk key: never renamed, never reused.
var psychicIntrusionMayCast = effects.S82OldB397MayCastPick("revealed-pick/psychic-intrusion-may-cast", "Psychic Intrusion")
