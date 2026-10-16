package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sting, the Glinting Dagger — Legendary Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+1 and has haste.
//	 At the beginning of each combat, untap equipped creature.
//	 Equipped creature has first strike as long as it's blocking or
//	 blocked by a Goblin or Orc.
//	 Equip {2}"
//
// The untap trigger is AtBeginningOfEachCombat (any player's turn,
// unlike AtBeginningOfYourCombat) — the effect reads the current host
// off the Dagger itself via attachedHostFor rather than off anything
// captured at Build, since which creature is equipped can change
// between combats.
//
// The conditional first strike is a GrantToAttachedWhile over live block
// state: the host is blocking an attacker that is a Goblin or an Orc, or
// is an attacker blocked by one (Card.BlockingTarget / AlsoBlocking, read
// through IsBlockingAttacker, so a creature that blocks several
// attackers counts each). DependsOnEvents lists EventBlock, which wakes
// the layer cache when blockers are declared and again when a block ends
// (combat ending, a blocker leaving combat), so the keyword is live at
// the first-strike damage step and gone after combat. Changelings count
// (Card.HasSubtype).
func init() {
	Register(Spec{
		OracleID:     "973c49a1-8425-40b9-8bdc-2c2434222314",
		Name:         "Sting, the Glinting Dagger",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("haste"),
			stingFirstStrike(),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfEachCombat("Sting, the Glinting Dagger — untap equipped creature", untapAttachedHost),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// stingFirstStrike is "Equipped creature has first strike as long as it's
// blocking or blocked by a Goblin or Orc."
func stingFirstStrike() game.StaticAbility {
	st := GrantToAttachedWhile(func(host *game.Card, g *game.Game, source *game.Card) bool {
		// "Goblin or Orc" are words of the source's text (CR 612.2, #3984).
		goblin, orc := source.TextSubtype("Goblin"), source.TextSubtype("Orc")
		isGoblinOrOrc := func(c *game.Card) bool { return c.HasSubtype(goblin) || c.HasSubtype(orc) }
		if g.Battlefield == nil {
			return false
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch {
			case host.IsBlockingAttacker(c.InstanceID) && isGoblinOrOrc(c):
				return true
			case c.IsBlockingAttacker(host.InstanceID) && isGoblinOrOrc(c):
				return true
			}
		}
		return false
	}, "first strike")
	st.Label = "Attached creature has first strike as long as it's blocking or blocked by a Goblin or Orc."
	st.DependsOnEvents = []game.EventKind{game.EventBlock}
	return st
}
