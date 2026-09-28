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
// The division is the caster's (#1563, CR 601.2d): the trigger's
// pick_target prompt asks for each target's share with the targets,
// every target at least 1 and the shares summing to 5. Each slot is
// re-checked at resolution on its own (CR 608.2b), so a creature that
// left in response takes nothing and the rest take exactly the shares
// announced for them — the departed target's share is lost, not
// redistributed.
func init() {
	Register(Spec{
		OracleID:        "daff70a4-a990-47c9-b6ef-2c379272c8a3",
		Name:            "Dragonlord Atarka",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetPermanent("any number of target creatures and/or planeswalkers your opponents control",
				And(Or(Creature(), Planeswalker()), OpponentControls())).WithCount(0, 5).Dividing(Divide(5)),
			Key: "Dragonlord Atarka — 5 damage divided among the targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
