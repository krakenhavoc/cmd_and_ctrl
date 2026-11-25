package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormchaser Chimera — Creature — Chimera {2}{U}{R}, 2/3 (Conspiracy: Take
// the Crown, S82 older sets batch 273, #5540):
//
//	"Flying
//	 {2}{U}{R}: Scry 1, then reveal the top card of your library. This
//	 creature gets +X/+0 until end of turn, where X is that card's mana
//	 value."
//
// Sifter Wurm's shape: the scry's continuation (a registered tail, so the
// prompt is a restore point) reveals the card that is now on top, after the
// scried card has been kept or bottomed, and reads its mana value as the
// library holds it (a land is 0; an empty library reveals nothing and gives
// +0/+0). The pump is pinned to the Chimera as this object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fccc57f6-2555-4ab3-b85a-83b524abe892",
		Name:            "Stormchaser Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{R}: Scry 1, then reveal the top card of your library. This creature gets +X/+0 until end of turn, where X is that card's mana value.",
			Purpose: game.Purpose{Scry: 1},
			Cost:    effects.ManaCost("{2}{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Scry{Player: ctx.Controller(), N: 1, ThenTail: ctx.ItemTail(stormchaserChimeraReveal)}.Apply(ctx)
			},
		}},
	})
}

// stormchaserChimeraReveal is the clause after the scry. An on-disk key:
// never renamed, never reused.
var stormchaserChimeraReveal = effects.CardTail("stormchaser-chimera/reveal-then-pump", func(ctx *effects.Context, _ game.TailOutcome) error {
	card, mv, err := effects.RevealTopReadingManaValue(ctx, "Stormchaser Chimera — reveal the top card of your library")
	if err != nil || card == uuid.Nil || mv <= 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return err
	}
	return effects.BoostUntilEOT{Target: ctx.Source(), Power: mv, Label: "Stormchaser Chimera — +X/+0"}.Apply(ctx)
})
