package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakish Scoundrel — Creature — Elf Rogue {2}{B}{G}, 3/3 (Murders at
// Karlov Manor, #4469):
//
//	"Deathtouch
//	 When this creature enters or is turned face up, target creature
//	 gains indestructible until end of turn.
//	 Disguise {4}{B/G}{B/G} (You may cast this card face down for {3} as
//	 a 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)"
//
// Offender at Large's one ability with two conditions (OnAny). Entering
// face down does not trigger it (a face-down permanent has no abilities,
// CR 708.3); turning it face up does (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d69cb63f-98cc-4699-a03d-4f1beeeb7065",
		Name:             "Rakish Scoundrel",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{B/G}{B/G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, Self,
					"Rakish Scoundrel — target creature gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return grantEachLegalTargetUntilEOT(NewContext(g, item), "indestructible",
							"Rakish Scoundrel — indestructible")
					}),
				TargetCreature("target creature")),
		},
	})
}
