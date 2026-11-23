package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titania, Voice of Gaea — Legendary Creature — Elemental {1}{G}{G}, 3/4
// (The Brothers' War, #5482):
//
//	"Reach
//	 Whenever one or more land cards are put into your graveyard from
//	 anywhere, you gain 2 life.
//	 At the beginning of your upkeep, if there are four or more land cards
//	 in your graveyard and you both own and control Titania, Voice of Gaea
//	 and a land named Argoth, Sanctum of Nature, exile them, then meld them
//	 into Titania, Gaea Incarnate."
//
// The life trigger is The Gitrog Monster's land-card arrival, made a batch
// trigger (OncePerBatch, CR 603.2c): lands milled or destroyed together are
// one occurrence and gain 2 life once. The upkeep meld is an intervening
// "if" (CR 603.4, ADR 0145), checked as it triggers and again as it
// resolves; MeldWith rechecks that both are still owned and controlled. The
// combined back face is Titania, Gaea Incarnate, its own record.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42bccda5-1754-4619-b94f-7a72f934a718",
		Name:            "Titania, Voice of Gaea",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB, game.EventMill}, effects.GitrogLandToGraveyard,
				"Titania, Voice of Gaea — you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2}))),
			effects.InterveningIf(effects.S82BroB20TitaniaMeldReady{},
				effects.AtYourUpkeep("Titania, Voice of Gaea — exile Titania and Argoth, Sanctum of Nature, then meld them into Titania, Gaea Incarnate",
					effects.Do(effects.MeldWith{Partner: "Argoth, Sanctum of Nature", PartnerType: "land"}))),
		},
	})
}
