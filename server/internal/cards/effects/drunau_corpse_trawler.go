package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "870ca989-0e72-4e74-9570-57e129298f2e",
		Name:         "Drunau Corpse Trawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Drunau Corpse Trawler — create a 2/2 black Zombie creature token",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Target Zombie gains deathtouch until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    ManaCost("{2}{B}"),
			Targets: TargetCreature("target Zombie", HasSubtype("Zombie")),
			Effect:  gainsDeathtouchUntilEndOfTurn,
		}},
	})
}

// gainsDeathtouchUntilEndOfTurn is "target creature gains deathtouch
// until end of turn" for each still-legal target.
func gainsDeathtouchUntilEndOfTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCards(item, g) {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
