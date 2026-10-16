package effects

// Thousand-Year Elixir — Artifact {3} (EDHREC rank 890, #3654):
//
//	"You may activate abilities of creatures you control as though those
//	 creatures had haste.
//	 {1}, {T}: Untap target creature."
//
// The static is Spec.ActivateAsThoughHaste (#3928): CR 302.6 no longer
// stops its controller paying a {T} or {Q} cost of a creature they
// control the turn it arrives, mana abilities included. It is an "as
// though" effect (CR 609.4), so the creatures gain no haste and still
// can't attack that turn; an opponent activating one of their "any
// player may activate" abilities is not "you" and gets nothing.
//
// The untap is the ordinary mana-plus-tap activated ability Voltaic Key
// has, over "target creature".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "4dc5726e-2f7e-4c2b-9616-c3301d212f78",
		Name:                  "Thousand-Year Elixir",
		Completeness:          CompletenessFull,
		ActivateAsThoughHaste: true,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target creature.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  untapFirstLegalTarget,
		}},
	})
}
