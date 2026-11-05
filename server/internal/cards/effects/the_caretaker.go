package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Caretaker — Legendary Creature — Elemental {5}{U}, 5/6 (Star Trek,
// #4973):
//
//	"Vigilance, ward {2}
//	 At the beginning of each end step, if you faced a dilemma this turn,
//	 target opponent exiles cards from the top of their library until
//	 they exile a nonland card. You may cast that card without paying its
//	 mana cost. (You face a dilemma as you choose one or more modes for a
//	 spell or ability.)"
//
// "If you faced a dilemma this turn" reads the per-turn count the engine
// keeps beside EventFacedDilemma (Game.DilemmasFacedThisTurn, cleared as
// each turn begins). It is an intervening if (CR 603.4): checked as the
// end step begins, and again as the ability resolves. It triggers at every
// player's end step, not only yours.
//
// The exile is the target opponent's library, until a nonland card
// (MillToZone's UntilNonland loop, which owns the mill-replacement and
// command-zone prompts), and the free cast is Lady Loki's and Nico
// Minoru's (exileUntilNonlandGranted): The Caretaker's controller may
// cast the nonland card during the resolution (CR 608.2g), for {0}, with
// an {X} at 0 (CR 107.3b). An uncast card and the lands exiled on the way
// stay in exile. A target that is no longer legal makes the ability do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f77ea393-4142-4c06-8bb3-589ca74a7338",
		Name:            "The Caretaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "The Caretaker — ward {2}"),
			Targeting(
				InterveningIf(CondFacedADilemmaThisTurn{},
					On(game.EventBeginEndStep, AllOf(AnyPlayer, YouFacedADilemmaThisTurn),
						"The Caretaker — target opponent exiles until a nonland card; you may cast it without paying its mana cost",
						theCaretakerExile)),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

// theCaretakerExile runs the exile off the target opponent's library,
// with the free cast riding the run's keyed tail.
func theCaretakerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		perm := exileUntilNonlandFreeCastPermission(ctx, "The Caretaker")
		tail := ctx.ItemTail(exileUntilNonlandGranted)
		tail.Permission = &perm
		return g.MillToZoneKeyedForEffect(t.ID, 0, game.ZoneExile, UntilNonland, tail)
	}
	return nil
}
