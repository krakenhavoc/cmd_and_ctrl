package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eris, Roar of the Storm — Legendary Creature — Elemental Warlock
// {8}{U}{R}, 4/4 (Outlaws of Thunder Junction Commander):
//
//	"This spell costs {2} less to cast for each different mana value among
//	 instant and sorcery cards in your graveyard.
//	 Flying, prowess
//	 Whenever you cast your second spell each turn, create a 4/4 red
//	 Dragon Elemental creature token with flying and prowess."
//
// The reduction is a SelfCostModifier priced at CR 601.2f off the graveyard
// as it stands (generic mana only, never below the coloured {U}{R}). The
// trigger is the second-spell clause Brimstone Roundup uses, off the
// per-turn cast tally; the cast of the Eris that is the second spell does
// not trigger itself, since it is on the stack, not the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "918991a6-7321-419d-8f3d-150fe32d02d2",
		Name:            "Eris, Roar of the Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return 2 * cmdr2S41DifferentInstantSorceryManaValues(q.Game, q.Controller)
			}, "This spell costs {2} less to cast for each different mana value among instant and sorcery cards in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Eris, Roar of the Storm — create a 4/4 red Dragon Elemental with flying and prowess",
				Do(CreateToken{Template: TokenCard("4/4 red Dragon Elemental with flying and prowess"), N: 1})),
		},
	})
}
