package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drunau Corpse Trawler — Creature — Zombie {3}{U}, 1/1 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"When this creature enters, create a 2/2 black Zombie creature token.
//	 {2}{B}: Target Zombie gains deathtouch until end of turn."
//
// An ordinary enters trigger around the shared Zombie token, and an
// activated pump aimed at any Zombie (the Trawler itself, the token, or
// another player's). A target that has left by resolution is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "870ca989-0e72-4e74-9570-57e129298f2e",
		Name:         "Drunau Corpse Trawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Drunau Corpse Trawler — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Target Zombie gains deathtouch until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    effects.ManaCost("{2}{B}"),
			Targets: effects.TargetCreature("target Zombie", effects.HasSubtype("Zombie")),
			Effect:  gainsDeathtouchUntilEndOfTurn,
		}},
	})
}

// gainsDeathtouchUntilEndOfTurn is "target creature gains deathtouch
// until end of turn" for each still-legal target.
func gainsDeathtouchUntilEndOfTurn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range effects.LegalTargetCards(item, g) {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
