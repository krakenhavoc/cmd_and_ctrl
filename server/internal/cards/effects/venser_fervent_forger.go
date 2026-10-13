package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venser, Fervent Forger — Legendary Creature — Human Sorcerer
// {4}{R}{R}, 5/3:
//
//	"Flash
//	 When Venser enters, choose one —
//	 • Copy target instant or sorcery spell an opponent controls twice.
//	   You may choose new targets for the copies.
//	 • Create two tokens that are copies of target permanent an
//	   opponent controls. They gain haste. At the beginning of the next
//	   end step, sacrifice them."
//
// Mode one is CopySpell with a count of two, controlled by you (CR
// 707.10). Mode two is Electroduplicate's token copy with haste and a
// scheduled sacrifice of both tokens. Each mode names its own target
// clause, so a mode with no legal target cannot be chosen.
//
// A token copy runs the ordinary battlefield entry (#762), so the copied
// card's as-enters clause fires for each token as well as its enters
// triggers; TestVenserTokenCopyAsksForTheCopiedCardsEntryChoice pins it.
func init() {
	venser := WhenThisEnters("Venser, Fervent Forger — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	venser.Modes = ChooseOne(
		ModeDoing("Copy target instant or sorcery spell an opponent controls twice. You may choose new targets for the copies.",
			instantOrSorcerySpell("target instant or sorcery spell an opponent controls", OpponentControls()),
			rfCreatureFCopyTargetSpellTwice),
		ModeDoing("Create two tokens that are copies of target permanent an opponent controls. They gain haste. At the beginning of the next end step, sacrifice them.",
			TargetPermanent("target permanent an opponent controls", OpponentControls()),
			rfCreatureFTwoHastyCopiesSacrificedAtEnd),
	)
	Register(Spec{
		OracleID:        "9b1ae6d8-0cf3-4a67-9241-37f4bc5ddbdd",
		Name:            "Venser, Fervent Forger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{venser},
	})
}
