package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmer — Enchantment {2}{U}{U} (Mirage, #5633):
//
//	"As this enchantment enters, choose a land type.
//	 Each land of the chosen type has phasing."
//
// The choice is CR 614.12's, made as it enters and stored on the permanent
// (Card.ChosenOption); each land type of CR 205.3i is its own layer 6 grant
// of phasing, gated on that option being the chosen one (ChosenIs), so until
// the choice is made no land phases. A granted phasing works like a printed
// one (game/phasing.go reads the layered keyword): every land of the type,
// every player's, phases out in its controller's next untap step and in the
// one after, before they untap.
//
// No simplification.
func init() {
	statics := make([]game.StaticAbility, 0, len(game.LandTypes))
	for _, t := range game.LandTypes {
		s := effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsLand() && target.HasSubtype(t)
		}, game.KeywordPhasing)
		s.Label = "Each land of the chosen type has phasing."
		statics = append(statics, effects.StaticWhenChosen(t, s))
	}
	effects.Register(effects.Spec{
		OracleID:     "f557fb2c-581b-4a8d-8766-1e46e92224fa",
		Name:         "Shimmer",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Shimmer", game.LandTypes...),
		Static:       statics,
	})
}
