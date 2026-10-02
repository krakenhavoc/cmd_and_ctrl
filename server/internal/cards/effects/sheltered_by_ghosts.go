package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheltered by Ghosts — Enchantment — Aura {1}{W}:
//
//	"Enchant creature you control
//	 When this Aura enters, exile target nonland permanent an opponent
//	 controls until this Aura leaves the battlefield.
//	 Enchanted creature gets +1/+0 and has lifelink and ward {2}."
//
// Ossification's exile "until this Aura leaves the battlefield"
// (CR 610.3, ExileUntil): the return is a one-shot effect performed
// immediately after the Aura leaves, never a trigger on the stack
// (#1729) — and it still happens if the Aura's owner leaves the game.
// The ward is the Aura's own trigger watching its host becoming a
// target (WardAttached).
//
// CR 610.3b, and the ruling of 2024-09-20: "If Sheltered by Ghosts
// leaves the battlefield before its triggered ability resolves, the
// target permanent won't be exiled at all." ExileUntil reads the Aura
// as the object it was when the ability triggered.
//
// The leave trigger the card carried before #1729 stays as a legacy
// row (UntilThisLeavesLegacyReturn), for a card exiled by an older
// binary.
//
// The permanent returns under its OWNER's control (CR 610.3c), as a new
// object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d13fc657-c6fc-4394-bc70-691050550226",
		Name:         "Sheltered by Ghosts",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetPermanent("target nonland permanent an opponent controls",
					Nonland(), OpponentControls()),
				Key:    shelteredByGhostsExileLabel,
				Effect: exileChosenTargetUntilThisLeaves("Sheltered by Ghosts — the exiled card returns when the Aura leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Sheltered by Ghosts — return the exiled card", shelteredByGhostsExileLabel),
			WardAttached(WardMana("{2}"), "Sheltered by Ghosts — ward {2}"),
		},
	})
}

// shelteredByGhostsExileLabel is the exile trigger's stack label; the
// "exiled with" record keys on it, so both halves share the constant.
const shelteredByGhostsExileLabel = "Sheltered by Ghosts — exile target nonland permanent an opponent controls"
