package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightcap Badger // Fungus Frolic — Creature — Badger Druid {3}{G},
// 3/4 // Instant — Adventure {2}{G}:
//
//	Brightcap Badger:
//	"Each Fungus and Saproling you control has "{T}: Add {G}."
//	 At the beginning of your end step, create a 1/1 green Saproling
//	 creature token."
//
//	Fungus Frolic:
//	"Create two 1/1 green Saproling creature tokens."
//
// The adventure layout is Brazen Borrower's: the creature under the bare
// oracle id, the Adventure under "<oracle_id>#1". The mana is a granted
// mana ability (ADR 0093) on every Fungus and Saproling creature you
// control, so a Saproling is subject to summoning sickness for its {T}
// like any creature (CR 302.6). The Badger is itself neither type.
//
// No simplification.
const cmdr2B17BadgerGrant = "brightcap-badger/tap-for-green"

func init() {
	const oracleID = "f61c0fc0-7241-455e-bbf5-f1f3862d46e3"

	Register(Spec{
		OracleID:     oracleID,
		Name:         "Brightcap Badger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Grants:       []AbilityGrant{TapForManaGrant(cmdr2B17BadgerGrant, "{G}", "Add {G}", "{T}: Add {G}.")},
		Static: []game.StaticAbility{
			Labelled(TribalAbilityGrant(TribeFilter{Tribes: []string{"Fungus", "Saproling"}, YoursOnly: true}, cmdr2B17BadgerGrant),
				"Each Fungus and Saproling you control has \"{T}: Add {G}.\""),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Brightcap Badger — create a 1/1 green Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})),
		},
	})

	Register(Spec{
		OracleID:     oracleID + "#1",
		Name:         "Fungus Frolic",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 2}.Apply(ctx)
		},
	})
}
