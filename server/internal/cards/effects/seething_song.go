package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seething Song — Instant for {2}{R}:
//
//	"Add {R}{R}{R}{R}{R}."
//
// Dark Ritual in red, two mana bigger and two mana better. A ritual,
// so its whole job is to turn three mana into five inside one step.
//
// This is the AddMana primitive's exact shape and the reason that
// primitive exists: "Add {R}{R}{R}{R}{R}" is a SPELL that resolves
// off the stack, so it cannot be a ManaAbility (CR 605.3b — mana
// abilities never use the stack), and until AddMana landed with the
// roadmap's batch 01 every mana in the engine came from
// ActivateManaAbility.
//
// The batch-02 triage (#295) filed Seething Song under "mana
// pipeline — no tracking issue yet". That was written before AddMana
// shipped; nothing new is needed for it now.
//
// The mana lands in the caster's pool attributed to the spell, and
// empties with the pool at the end of the step (CR 106.4). That is
// not a limitation, it is the card: a ritual you cast and then fail
// to spend is a ritual you wasted, exactly as in paper.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "64bf8929-f5f2-4d50-8667-13b1d007bcfc",
		Name:         "Seething Song",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddMana{Player: item.Controller, Produced: "{R}{R}{R}{R}{R}"}.Apply(ctx)
		},
	})
}
