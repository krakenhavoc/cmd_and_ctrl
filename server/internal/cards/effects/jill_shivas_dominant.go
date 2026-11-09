package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jill, Shiva's Dominant // Shiva, Warden of Ice (Final Fantasy, #4462).
// Front, Legendary Creature — Human Noble Warrior {2}{U}, 2/2:
//
//	"When Jill enters, return up to one other target nonland permanent
//	 to its owner's hand.
//	 {3}{U}{U}, {T}: Exile Jill, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery."
//
// Back, Legendary Enchantment Creature — Saga Elemental, 4/5:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I, II — Mesmerize — Target creature can't be blocked this turn.
//	 III — Cold Snap — Tap all lands your opponents control. Exile
//	 Shiva, then return it to the battlefield (front face up)."
//
// Dion's flip. Jill's enter is an "up to one" bounce of a nonland
// permanent that is not Jill. Mesmerize is Artful Dodge's can't-be-blocked
// on any target creature. Cold Snap taps every land each opponent controls
// as it resolves, then exiles Shiva and returns it front face up (Jill
// enters again, bouncing another permanent).
//
// No simplification.
func init() {
	upToOne := TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1)
	upToOne.ExcludeSource = true
	mesmerize := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Shiva, Warden of Ice — Mesmerize: can't be blocked"}.Apply(ctx)
	}
	Register(Spec{
		OracleID:     jillOracleID,
		Name:         "Jill, Shiva's Dominant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Jill, Shiva's Dominant — return up to one other target nonland permanent to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BounceToHand{Target: id}.Apply(ctx)
				}), upToOne),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{U}{U}, {T}: Exile Jill, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{U}{U}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     jillOracleID + "#1",
		Name:         "Shiva, Warden of Ice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Shiva, Warden of Ice", 1, "Mesmerize — target creature can't be blocked this turn"), TargetCreature("target creature"), mesmerize),
			ChapterTriggerTargeting(2, SagaChapterLabel("Shiva, Warden of Ice", 2, "Mesmerize — target creature can't be blocked this turn"), TargetCreature("target creature"), mesmerize),
			ChapterTrigger(3, SagaChapterLabel("Shiva, Warden of Ice", 3, "Cold Snap — tap all lands your opponents control; exile Shiva, then return it front face up"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := b24TapAllLandsControlledBy(ctx, opp); err != nil {
							return err
						}
					}
					return ChapterExileAndReturnTransformed(g, item)
				}),
		},
	})
}

// jillOracleID is shared by both faces.
const jillOracleID = "c6655e05-9e31-4a74-a97d-1eeb473fe900"
