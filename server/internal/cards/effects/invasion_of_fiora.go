package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Fiora // Marchesa, Resolute Monarch — March of the Machine
// (#5652, part of #5477), a transforming double-faced card:
//
//	Invasion of Fiora — Battle — Siege {4}{B}{B}, defense 4
//	  "(As a Siege enters, choose an opponent to protect it. You and
//	    others can attack it. When it's defeated, exile it, then cast it
//	    transformed.)
//	   When this Siege enters, choose one or both —
//	   • Destroy all legendary creatures.
//	   • Destroy all nonlegendary creatures."
//	Marchesa, Resolute Monarch — Legendary Creature — Human Noble, 3/6
//	  "Menace, deathtouch
//	   Whenever Marchesa attacks, remove all counters from up to one
//	   target permanent.
//	   At the beginning of your upkeep, if you haven't been dealt combat
//	   damage since your last turn, you draw a card and you lose 1 life."
//
// The Siege under the bare oracle ID, Marchesa under "<oracle>#1". The
// protector, the defense counters and the defeat are the engine's, keyed
// on the card type (game/battle.go).
//
// With both modes chosen they happen in the printed order: the legendary
// creatures are destroyed, then the nonlegendary ones, each a destruction
// of its own. "Legendary" is read as each mode is carried out.
//
// Marchesa's upkeep "if" is the engine's since-last-turn memory
// (game/since_last_turn.go, #5652), asked as the upkeep begins and again
// as the ability resolves (CR 603.4). "Remove all counters from up to one
// target" is Purging Stormbrood's body, over any permanent: every kind it
// carries, if the target is still legal (CR 608.2b).
//
// No simplification.

const invasionOfFioraOracleID = "7209f2b6-8851-42ab-bac6-5cf06e211e46"

func init() {
	enters := WhenThisEnters("Invasion of Fiora — choose one or both", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOneOrMore(
		ModeDoing("Destroy all legendary creatures.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return DestroyAllMatching{Match: And(Creature(), Legendary())}.Apply(ctx)
		}),
		ModeDoing("Destroy all nonlegendary creatures.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return DestroyAllMatching{Match: And(Creature(), Not(Legendary()))}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     invasionOfFioraOracleID,
		Name:         "Invasion of Fiora",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(enters, game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
			DefeatedTrigger("Invasion of Fiora — defeated: exile it, then cast Marchesa, Resolute Monarch", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfFioraOracleID + "#1",
		Name:            "Marchesa, Resolute Monarch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Marchesa, Resolute Monarch — remove all counters from up to one target permanent", purgingStormbroodRemoveCounters),
				TargetPermanent("up to one target permanent").WithCount(0, 1)),
			AtYourUpkeepIfNoCombatDamageSinceYourLastTurn("Marchesa, Resolute Monarch — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
