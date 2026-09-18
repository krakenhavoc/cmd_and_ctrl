package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aetheric Amplifier — Artifact {3} (EDHREC rank 2428):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Choose one. Activate only as a sorcery.
//	 • Double the number of each kind of counter on target permanent.
//	 • Double the number of each kind of counter you have."
//
// A Manalith that grows into a Deepglow Skate. The mana ability is
// the Birds shape — any colour, the printed width, so the pipe is not
// narrowed to the commander's identity. The activation is a CR 602
// ability at sorcery speed sharing the tap.
//
// #764 made it modal: ActivatedAbility.Modes is the same
// game.ModeSpec a modal spell declares, and the choice is announced
// at ACTIVATION with the targets (CR 602.2b) — one indivisible step,
// no prompt, because the player who activates is the player who
// chooses. The first bullet is Deepglow Skate's per-target doubler
// (each kind snapshotted first, so a Doubling Season firing on one
// kind cannot change what the next receives) on one target permanent,
// anyone's, as printed; the second doubles the activator's own player
// counters — poison, energy, experience, rad — and targets nothing,
// so activating it offers no picker at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "295cd8e1-0830-46c1-9957-556afd4bcee6",
		Name:         "Aetheric Amplifier",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Choose one — double the counters on target permanent; or double the counters you have.",
			Cost:         Plus(ManaCost("{4}"), TapCost()),
			SorcerySpeed: true,
			Modes: ChooseOne(
				ModeDoing("Double the number of each kind of counter on target permanent.",
					TargetPermanent("target permanent"),
					func(item *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return b17DoubleCountersOn(ctx, t.ID)
					}),
				ModeDoing("Double the number of each kind of counter you have.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return b64DoublePlayerCounters(ctx, item.Controller)
					}),
			),
		}},
	})
}

// b64DoublePlayerCounters doubles every kind of player counter the
// given player has — poison, energy, experience, rad. The counts are
// snapshotted before the first addition, for the same reason
// b17DoubleCountersOn snapshots a permanent's: a replacement that
// fires on one kind must not change what the next kind receives.
func b64DoublePlayerCounters(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil {
		return nil
	}
	add := make(map[string]int, len(p.Counters))
	for kind, n := range p.Counters {
		if n > 0 {
			add[kind] = n
		}
	}
	for kind, n := range add {
		if err := ctx.Game.AddPlayerCounterForEffect(player, kind, n); err != nil {
			return err
		}
	}
	return nil
}
