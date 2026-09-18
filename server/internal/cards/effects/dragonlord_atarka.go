package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonlord Atarka — Legendary Creature — Elder Dragon {5}{R}{G},
// 8/8 (EDHREC rank 4540):
//
//	"Flying, trample
//	 When Dragonlord Atarka enters, it deals 5 damage divided as you
//	 choose among any number of target creatures and/or planeswalkers
//	 your opponents control."
//
// Seven mana for an 8/8 flying trampler that sweeps five points off
// the table on the way in. In a Gruul or Jund deck she is a one-card
// swing: the ETB answers two or three blockers and the body ends the
// game two turns later. She is also a commander in her own right,
// where the damage repeats every cast.
//
// "YOUR OPPONENTS CONTROL" is the one-sided clause and it is enforced
// on the target slot, so your own board is never a legal pick. "Any
// number" is bounded above by five, because every chosen target has
// to be assigned at least one damage (CR 601.2d) — one to five slots,
// all distinct.
//
// DECLARED SIMPLIFICATION (weaker than printed): the DIVISION is made
// for the player. The engine's target picker for a TRIGGERED ability
// carries no damage distribution — game.StackItem.Distribution exists
// for a cast, not for the harvester's pick_target prompt — so the five
// points are split as evenly as possible across the chosen targets in
// the order they were picked, the remainder going to the earliest
// picks: one target takes 5, two take 3/2, three take 2/2/1, five take
// one each. A 4/1 split is not offered. This is Fury's caveat exactly,
// and the seam is the same one: a distribution on the pick_target
// prompt.
//
// Each slot is re-checked at resolution on its own (CR 608.2b), so a
// creature that left in response is skipped and the rest are still
// hit — the points assigned to the departed target are simply not
// dealt, which is weaker than the printed card (where the division is
// locked in on announce and the remaining targets keep their share)
// and never stronger.
func init() {
	Register(Spec{
		OracleID:     "daff70a4-a990-47c9-b6ef-2c379272c8a3",
		Name:         "Dragonlord Atarka",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The 5 damage is divided as evenly as possible among the targets you pick, in the order you pick them, rather than however you choose.",
		},
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetPermanent("any number of target creatures and/or planeswalkers your opponents control",
				And(Or(Creature(), Planeswalker()), OpponentControls())).WithCount(0, 5),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dragonlord Atarka — 5 damage divided among the targets",
					func(g *game.Game, item *game.StackItem) error {
						return b22DamageDividedEvenly(NewContext(g, item), 5)
					})
			},
		}},
	})
}
