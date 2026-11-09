package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Osteomancer Adept — Creature — Squirrel Warlock {1}{B}, 2/2
// (Bloomburrow, #4467):
//
//	"Deathtouch
//	 {T}: Until end of turn, you may cast creature spells from your
//	 graveyard by foraging in addition to paying their other costs. If you
//	 cast a spell this way, that creature enters with a finality counter on
//	 it. (To forage, exile three cards from your graveyard or sacrifice a
//	 Food. If a creature with a finality counter on it would die, exile it
//	 instead.)"
//
// The ability grants two standing cast permissions to its controller until
// end of turn, over creature cards in their own graveyard (CastPermission,
// ADR 0066), one for each way to forage: AltCostKey osteomancer-forage-exile
// charges the card's printed mana cost and three OTHER cards of the
// graveyard exiled (ExileOtherFromGraveyard), osteomancer-forage-food the
// printed cost and a Food sacrificed (SacrificeFood). Both are marked Forage,
// so paying either emits EventForage as any forage does (CR 701.61a;
// "whenever you forage" sees it), and a graveyard with fewer than three
// other cards and no Food opens neither. The cost is an additional one, so
// the spell's own timing, targets and other costs are untouched (it is
// cast at sorcery speed unless it has flash, the ruling). The finality
// counter is the permission's entry rider (EntersWithCounterKind, CR
// 614.1c); the engine's CR 122.1h replacement does the rest. Dryad Arbor is
// a land and is played, never cast, so it is not opened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "791cbb7c-7935-4902-b050-9ea9d030e9fa",
		Name:            "Osteomancer Adept",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Until end of turn, you may cast creature spells from your graveyard by foraging in addition to paying their other costs. If you cast a spell this way, that creature enters with a finality counter on it.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, p := range []game.CastPermission{
					{
						AltCostKey:              "osteomancer-forage-exile",
						ExileOtherFromGraveyard: game.ForageExileCount,
						Label:                   "Cast a creature spell from your graveyard by foraging (exile three other cards from your graveyard); it enters with a finality counter (Osteomancer Adept)",
					},
					{
						AltCostKey:    "osteomancer-forage-food",
						SacrificeFood: true,
						Label:         "Cast a creature spell from your graveyard by foraging (sacrifice a Food); it enters with a finality counter (Osteomancer Adept)",
					},
				} {
					p.Player = ctx.Controller()
					p.Zone = game.ZoneGraveyard
					p.Scope = game.ScopeStanding
					p.CastOnly = true
					p.Forage = true
					p.Filter = game.PermissionFilter{CreatureOnly: true}
					p.EntersWithCounterKind = game.CounterFinality
					p.EntersWithCounters = 1
					p.Source = ctx.Source()
					g.GrantCastPermissionForEffect(p)
				}
				return nil
			},
		}},
	})
}
