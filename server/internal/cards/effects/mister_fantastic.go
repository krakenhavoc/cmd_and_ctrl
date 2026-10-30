package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mister Fantastic — Legendary Creature — Human Scientist Hero {2}{U},
// 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Reach, vigilance
//	 At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, draw a card.
//	 {R}{G}{W}{U}, {T}: Copy target triggered ability you control twice. You
//	 may choose new targets for the copies."
//
// The draw is an intervening if (CR 603.4) on the per-turn cast tally. The
// activated ability is Strionic Resonator's target clause (a triggered
// ability on the stack that its controller controls) and copy primitive,
// asked for two copies (CopyAbility.Count): each copy remembers the
// original's triggering event (CR 707.10) and offers its own new targets
// (CR 707.10c). The tap waits out summoning sickness, as a creature's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a8011dce-17fb-46e1-8297-81a4acccbd05",
		Name:            "Mister Fantastic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			mscM13CombatIfCastNoncreature("Mister Fantastic — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}{G}{W}{U}, {T}: Copy target triggered ability you control twice. You may choose new targets for the copies.",
			Cost:    Plus(ManaCost("{R}{G}{W}{U}"), TapCost()),
			Targets: AbilityOnStack("target triggered ability you control", AnAbilityYouControl(), TriggeredAbilityOnly()),
			Effect:  misterFantasticCopyTwice,
		}},
	})
}

func misterFantasticCopyTwice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CopyAbility{ItemID: id, Controller: item.Controller, Count: 2, ChooseNewTargets: true}.Apply(ctx)
}
