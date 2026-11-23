package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kagha, Shadow Archdruid — {2}{B}{G} Legendary Creature — Elf Druid 1/4
//
//	"Whenever Kagha attacks, it gains deathtouch until end of turn. Mill two
//	 cards.
//	 Once during each of your turns, you may play a land or cast a permanent
//	 spell from among cards in your graveyard that were put there from your
//	 library this turn."
//
// The second line is a standing graveyard permission (ADR 0066, Serra
// Paragon's one-use-per-turn budget over a land or a permanent spell) on your
// own turn, narrowed by PermissionFilter.FromLibraryToGraveyardThisTurn to
// cards that arrived from a library this turn: the cards the attack trigger
// milled, a surveil's, or any other move out of a library. A card that came
// from a hand or the battlefield, or that left the graveyard and came back
// (CR 400.7), does not qualify. A land played this way is a land play.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f78db9ad-d4a3-4e49-a76e-bd31124b7837",
		Name:         "Kagha, Shadow Archdruid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kagha, Shadow Archdruid — it gains deathtouch until end of turn; mill two cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"deathtouch"}, Label: "Kagha, Shadow Archdruid — deathtouch"}).Apply(ctx); err != nil {
						return err
					}
					return effects.MillCards{N: 2}.Apply(ctx)
				}),
		},
		CastPermissions: []game.CastPermission{{
			Zone:    game.ZoneGraveyard,
			Filter:  game.PermissionFilter{CardInZone: true, FromLibraryToGraveyardThisTurn: true, PermanentOnly: true},
			PerType: []string{game.PermissionTypePermanent},
			Timing:  game.TimingYourTurnOnly,
			Label:   "Once during each of your turns, play a land or cast a permanent spell from among cards in your graveyard that were put there from your library this turn (Kagha, Shadow Archdruid)",
		}},
	})
}
