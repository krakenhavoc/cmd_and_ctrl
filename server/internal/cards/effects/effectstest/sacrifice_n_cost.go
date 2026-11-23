package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_n_cost_test.go — #747 (ADR 0020 addendum §12–§15, ADR
// 0021 addendum): a sacrifice cost of N permanents, at all three cost
// sites — an activated ability, a mana ability, and an additional cost
// to cast — through fixture specs, so the engine rules are pinned
// apart from any one card. The cards that use it are in
// sacrifice_n_cards_test.go.

const (
	SacNAbilityOracle  = "test-sacrifice-n-ability"
	SacNSelfTooOracle  = "test-sacrifice-n-self-too"
	SacNManaOracle     = "test-sacrifice-n-mana"
	SacNSpellOracle    = "test-sacrifice-n-spell"
	SacNWatcherOracle  = "test-sacrifice-n-dies-watcher"
	SacSelfOtherOracle = "test-sacrifice-self-and-one-other"
)

func init() {
	// "Sacrifice two creatures: You gain 1 life."
	effects.Register(effects.Spec{
		OracleID: SacNAbilityOracle,
		Name:     "Two-Creature Outlet",
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two creatures: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(2, "two creatures", effects.Creature()),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
	// "{T}, Sacrifice this and two other creatures: You gain 1 life."
	// The source is itself a creature, so naming it among the two is
	// the "source named twice" case.
	effects.Register(effects.Spec{
		OracleID: SacNSelfTooOracle,
		Name:     "Self And Two Outlet",
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this and two creatures: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.SacrificeThis(), effects.SacrificeN(2, "two creatures", effects.Creature())),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
	// "Sacrifice two creatures: Add {C}{C}{C}."
	effects.Register(effects.Spec{
		OracleID: SacNManaOracle,
		Name:     "Two-Creature Altar",
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeN(2, "two creatures", effects.Creature()).SacrificeOther},
			Produced: "{C}{C}{C}",
			Label:    "Sacrifice two creatures: Add {C}{C}{C}",
			Answers:  game.AnswerSacOutlet,
		}},
	})
	// "As an additional cost to cast this spell, sacrifice two
	// creatures. Draw two cards."
	effects.Register(effects.Spec{
		OracleID:       SacNSpellOracle,
		Name:           "Two-Creature Rites",
		AdditionalCost: effects.SacrificeNCost(2, "two creatures", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{N: 2}.Apply(ctx)
		},
	})
	// "Whenever this or another creature dies, you gain 1 life." An
	// untargeted Blood Artist, so a test can count the triggers by
	// life total without answering prompts.
	effects.Register(effects.Spec{
		OracleID: SacNWatcherOracle,
		Name:     "Dies Watcher",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Dies Watcher — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
	// "Whenever this or another creature dies, you gain 1 life.
	// Sacrifice this and a creature: no effect." The pre-#747
	// two-permanent shape — SacrificeSelf plus a one-permanent
	// SacrificeOther — on a card that is its own dies-watcher.
	effects.Register(effects.Spec{
		OracleID: SacSelfOtherOracle,
		Name:     "Self And One Outlet",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Self And One Outlet — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this and a creature: no effect.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.SacrificeThis(), effects.SacrificeACreature()),
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}

func CountEventsOf(g *game.Game, kind game.EventKind, ids ...uuid.UUID) int {
	want := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && want[ev.CardID] {
			n++
		}
	}
	return n
}

func CardNameOf(g *game.Game, id uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(id); ok {
		return c.Name
	}
	return id.String()
}
