package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Static Prison — Enchantment {W}:
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield.
//	 You get {E}{E} (two energy counters).
//	 At the beginning of your first main phase, sacrifice this
//	 enchantment unless you pay {E}."
//
// ADR 0129 §3 (#1995). The exile is Ossification's "until this leaves
// the battlefield" (CR 610.3): one exile, and the return performed as
// the Prison leaves, before anyone gets priority; a Prison removed
// before its trigger resolves exiles nothing (CR 610.3b), and its
// controller still gets the energy. The upkeep of the Prison is CR
// 118.12a's pay-unless with an energy payment: declining, or being
// short (CR 118.3), sacrifices it, which returns the exiled card.
//
// No simplification.
func init() {
	exile := exileChosenTargetUntilThisLeaves("Static Prison — the exiled card returns when Static Prison leaves the battlefield")
	Register(Spec{
		OracleID:     "b1bc8e58-ea4f-46a9-b833-d780cf6758b3",
		Name:         "Static Prison",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
				Key:       "Static Prison — exile target nonland permanent an opponent controls until it leaves; you get {E}{E}",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := exile(g, item); err != nil {
						return err
					}
					return GetEnergy{N: 2}.Apply(NewContext(g, item))
				},
			},
			AtYourPrecombatMain("Static Prison — sacrifice it unless you pay {E}",
				sacrificeThisUnlessYouPayEnergy("Static Prison", "Static Prison", 1)),
		},
	})
}
