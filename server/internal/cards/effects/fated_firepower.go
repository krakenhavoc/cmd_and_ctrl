package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fated Firepower — Enchantment {X}{R}{R}{R} (EDHREC rank 2985):
//
//	"Flash
//	 This enchantment enters with X fire counters on it.
//	 If a source you control would deal damage to an opponent or a
//	 permanent an opponent controls, it deals that much damage plus
//	 an amount of damage equal to the number of fire counters on this
//	 enchantment instead."
//
// The Duskmourn damage amplifier. Flash rides PrintedKeywords (the
// cast path reads it off a card in hand). The fire counters go on as
// the spell resolves (Goldvein Hydra's posture, declared below), and
// the replacement adds the count read AT THE MOMENT THE DAMAGE WOULD
// BE DEALT — a proliferate that lands after the Firepower grows every
// later hit — to any damage from a source the controller controls
// aimed at an opponent or one of their permanents: combat damage, a
// Bolt, a pinger, a fight (b28AddCountersToDamageAtOpponents). Damage
// to the controller's own permanents, or to themselves, is untouched,
// as printed. With zero fire counters the replacement does not apply
// at all. CR 616: alongside a doubler the affected player orders the
// two, so "+N then ×2" and "×2 then +N" are both reachable.
//
// One declared simplification, weaker than printed: the X fire
// counters are placed as the spell resolves, a beat before the card
// enters (an entry replacement cannot read the spell's X), so a
// "whenever you put counters on a permanent" payoff does not see
// them.
func init() {
	Register(Spec{
		OracleID:        "13daa21c-278d-45bd-9a6e-a77d6a558453",
		Name:            "Fated Firepower",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The X fire counters are put on the enchantment as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		PrintedKeywords: []string{"flash"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "fire", N: ctx.X()}.Apply(ctx)
		},
		Replacements: []game.ReplacementEffect{
			b28AddCountersToDamageAtOpponents("Fated Firepower: extra damage for each fire counter", "fire"),
		},
	})
}
