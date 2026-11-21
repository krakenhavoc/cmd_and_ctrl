package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const LifeOfToshiroUmezawaOracleID = "0a708566-2994-47b3-8d89-0547cc115e7c"

func init() {
	const name = "Life of Toshiro Umezawa"
	chapter := func(n int) game.TriggeredAbility {
		t := effects.ChapterTrigger(n, effects.SagaChapterLabel(name, n, "choose one: +2/+2, -1/-1, or gain 2 life"),
			func(*game.Game, *game.StackItem) error { return nil })
		t.Modes = effects.ChooseOne(
			effects.ModeDoing("Target creature gets +2/+2 until end of turn.", effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					tgt, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: tgt.ID, Power: 2, Toughness: 2, Label: name + " — +2/+2"}.Apply(ctx)
				}),
			effects.ModeDoing("Target creature gets -1/-1 until end of turn.", effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					tgt, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: tgt.ID, Power: -1, Toughness: -1, Label: name + " — -1/-1"}.Apply(ctx)
				}),
			effects.ModeDoing("You gain 2 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Amount: 2}.Apply(ctx)
				}),
		)
		return t
	}
	effects.Register(effects.Spec{
		OracleID:     LifeOfToshiroUmezawaOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1), chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     LifeOfToshiroUmezawaOracleID + "#1",
		Name:         "Memory of Toshiro",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced:     "{B}",
			Restrictions: []string{effects.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			Label:        "{T}, Pay 1 life: Add {B}. Spend this mana only to cast an instant or sorcery spell",
		}},
	})
}
