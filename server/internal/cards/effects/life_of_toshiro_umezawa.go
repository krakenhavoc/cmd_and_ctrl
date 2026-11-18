package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Life of Toshiro Umezawa // Memory of Toshiro — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {1}{B}:
//
//	"I, II — Choose one —
//	 • Target creature gets +2/+2 until end of turn.
//	 • Target creature gets -1/-1 until end of turn.
//	 • You gain 2 life.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Human Samurai, 2/3:
//
//	"{T}, Pay 1 life: Add {B}. Spend this mana only to cast an instant or
//	 sorcery spell."
//
// The mode is chosen each time the chapter goes on the stack (CR 700.2), so
// chapters I and II may pick differently. The mana ability is Hydro-
// Channeler's restriction pair with a life cost; summoning sickness gates the
// tap (CR 302.6).
//
// No simplification.
const lifeOfToshiroUmezawaOracleID = "0a708566-2994-47b3-8d89-0547cc115e7c"

func init() {
	const name = "Life of Toshiro Umezawa"
	chapter := func(n int) game.TriggeredAbility {
		t := ChapterTrigger(n, SagaChapterLabel(name, n, "choose one: +2/+2, -1/-1, or gain 2 life"),
			func(*game.Game, *game.StackItem) error { return nil })
		t.Modes = ChooseOne(
			ModeDoing("Target creature gets +2/+2 until end of turn.", TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					tgt, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: tgt.ID, Power: 2, Toughness: 2, Label: name + " — +2/+2"}.Apply(ctx)
				}),
			ModeDoing("Target creature gets -1/-1 until end of turn.", TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					tgt, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: tgt.ID, Power: -1, Toughness: -1, Label: name + " — -1/-1"}.Apply(ctx)
				}),
			ModeDoing("You gain 2 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Amount: 2}.Apply(ctx)
				}),
		)
		return t
	}
	Register(Spec{
		OracleID:     lifeOfToshiroUmezawaOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1), chapter(2),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:     lifeOfToshiroUmezawaOracleID + "#1",
		Name:         "Memory of Toshiro",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Life: 1},
			Produced:     "{B}",
			Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			Label:        "{T}, Pay 1 life: Add {B}. Spend this mana only to cast an instant or sorcery spell",
		}},
	})
}
