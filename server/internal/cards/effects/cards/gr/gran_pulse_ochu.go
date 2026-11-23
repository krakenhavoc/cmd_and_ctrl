package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gran Pulse Ochu — Creature — Plant Beast {G}, 1/1 (Final Fantasy,
// #4462):
//
//	"Deathtouch
//	 {8}: Until end of turn, this creature gets +1/+1 for each permanent
//	 card in your graveyard."
//
// Azula, Ruthless Firebender's counted pump: the count is taken once,
// as the ability resolves (CR 608.2h), from the descend count of
// permanent cards (CR 110.4) in its controller's graveyard, and the
// bonus does not change if the graveyard does afterwards. A source that
// has left the battlefield gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "09b71c95-7b5d-4cee-a505-e0dda72715b7",
		Name:            "Gran Pulse Ochu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Until end of turn, this creature gets +1/+1 for each permanent card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{8}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				if !effects.OnBattlefield(g, self) {
					return nil
				}
				n := effects.PermanentCardsInGraveyard(g, item.Controller)
				return effects.BoostUntilEOT{Target: self, Power: n, Toughness: n,
					Label: "Gran Pulse Ochu — +1/+1 for each permanent card in your graveyard"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
