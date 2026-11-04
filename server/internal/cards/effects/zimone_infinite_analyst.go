package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zimone, Infinite Analyst — Legendary Creature — Human Wizard
// {1}{G}{U}, 0/4 (Secrets of Strixhaven Commander, #4834):
//
//	"The first spell you cast with {X} in its mana cost each turn costs
//	 {1} less to cast for each +1/+1 counter on Zimone.
//	 Whenever you cast your first spell with {X} in its mana cost each
//	 turn, put two +1/+1 counters on Zimone."
//
// Animar's reduction (a cost modifier counting the +1/+1 counters on the
// source, generic mana only) behind Lattice Library's first-X-spell
// condition. Both halves read the turn's cast tally (CastTally.XSpells),
// so a spell with {X} cast before Zimone arrived makes the next one not
// the first. The reduction is priced before the spell is cast and the
// counters arrive after, so the first X spell is cheaper by the counters
// Zimone already had and the next by the two it just got, and the
// reduction never touches the colored requirements (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "977f9390-e01d-4e9d-8e9c-e543b1266972",
		Name:         "Zimone, Infinite Analyst",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLessEach(CountersOnSource(game.CounterPlusOne),
				"The first spell you cast with {X} in its mana cost each turn costs {1} less to cast for each +1/+1 counter on Zimone.",
				YourSpell(), socB16FirstXSpellYouCastThisTurn()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, socB9FirstXSpellThisTurn,
				"Zimone, Infinite Analyst — put two +1/+1 counters on Zimone",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
