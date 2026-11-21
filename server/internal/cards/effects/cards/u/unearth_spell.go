package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unearth (the card) — Sorcery {B} (EDHREC top 1,000, #3654):
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Zombify with a mana value cap, plus the cycling ability every cycling
// card carries (cycling.go). The target is legal only while the card
// still has mana value 3 or less in the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e0b647a-6e1a-421b-bca9-8d5e346bda80",
		Name:         "Unearth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", effects.Creature(), effects.ManaValueLE(3), effects.YouOwn()),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
