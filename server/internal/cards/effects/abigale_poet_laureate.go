package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abigale, Poet Laureate // Heroic Stanza — Legendary Creature — Bird Bard
// {1}{W}{B}, 2/3, with the prepare spell Heroic Stanza, Sorcery {1}{W/B}
// (Secrets of Strixhaven, #4456):
//
//	"Flying
//	 Whenever you cast a creature spell, Abigale becomes prepared. (While
//	 it's prepared, you may cast a copy of its spell. Doing so unprepares
//	 it.)"
//
//	Heroic Stanza: "Put a +1/+1 counter on target creature."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". A creature spell cast while Abigale is on
// the battlefield prepares her (the spell is already cast when the trigger
// is raised, and a prepare-spell copy is a sorcery, not a creature spell);
// one that is already prepared stays so (BecomePrepared).
//
// No simplification.
func init() {
	const oracle = "2f5f46ed-b8aa-4864-bd20-17281d4632bf"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Abigale, Poet Laureate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Abigale, Poet Laureate — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Heroic Stanza",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    heroicStanzaResolve,
	})
}

// heroicStanzaResolve puts a +1/+1 counter on the target.
func heroicStanzaResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
