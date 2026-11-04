package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Campus Composer // Aqueous Aria — a preparation card (CR 722, ADR 0090),
// Secrets of Strixhaven (#4456):
//
//	Campus Composer {3}{U}, Creature — Merfolk Bard 3/4:
//	"Ward {2}
//	 This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Aqueous Aria {4}{U}, Sorcery:
//	"Create a 3/3 blue and red Elemental creature token with flying."
//
// Ward is the shared Ward trigger; "enters prepared" is SelfEntersPrepared.
//
// No simplification.
func init() {
	const id = "9b3155c1-c662-4ff1-80e2-4064e6233dc5"
	Register(Spec{
		OracleID:     id,
		Name:         "Campus Composer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Campus Composer — ward {2}"),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Aqueous Aria",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("3/3 blue and red Elemental with flying"), N: 1}.Apply(ctx)
		},
	})
}
