package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Innkeeper's Talent — Enchantment — Class {1}{G} (EDHREC rank 692):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control.
//	 {G}: Level 2
//	 Permanents you control with counters on them have ward {1}.
//	 {3}{G}: Level 3
//	 If you would put one or more counters on a permanent or player,
//	 put twice that many of each of those kinds of counters on that
//	 permanent or player instead."
//
// ADR 0071 Class machinery: LevelUp for each level, and each level's
// ability gated with ActiveWhen, so it exists only from that level on
// (CR 716.2a).
//
//   - Level 1 is an ordinary targeted beginning-of-combat trigger.
//   - Level 2 is a granted ward (WardGranted, Flowering of the White
//     Tree's shape): ward is a triggered ability with a cost, so it is
//     granted as the Class's trigger watching for one of your
//     permanents with a counter on it becoming the target of an
//     opponent's spell or ability. "Counters" is any kind.
//   - Level 3 is Vorinclex, Monstrous Raider's doubling half: a CR 614
//     replacement keyed on who PUTS the counters, on a permanent or a
//     player, so proliferate and an opponent's permanent you put
//     counters on are doubled too, and counters an opponent puts on
//     your permanent are not.
//
// The engine records who puts every placement (#3942), so a cost you
// pay, a permanent of yours entering with counters and your Saga's
// lore counter are all doubled, and none of an opponent's is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "915b75d1-b5e3-4aca-a54c-beaccef0a2e4",
		Name:         "Innkeeper's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Innkeeper's Talent — put a +1/+1 counter on target creature you control",
				PlusOneCounterOnChosenTarget),
				TargetCreature("target creature you control", YouControl())),
			AtLevel(2, WardGranted(WardMana("{1}"), "Innkeeper's Talent — permanents you control with counters on them have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && hasCounterOfKind(*target, "")
				})),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{G}")),
			LevelUp(3, ManaCost("{3}{G}")),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				return counterPlacementBy(ev, g, src.Controller, true)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:      "Innkeeper's Talent: double the counters you put",
			ActiveWhen: Level(3),
		}},
	})
}
