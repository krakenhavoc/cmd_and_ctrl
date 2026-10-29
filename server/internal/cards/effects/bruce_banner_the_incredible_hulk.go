package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bruce Banner // The Incredible Hulk — a double-faced Legendary Creature
// (Marvel Super Heroes, #4455):
//
//	Bruce Banner {U}, Legendary Creature — Human Scientist Hero, 1/1
//	  "{X}{X}, {T}: Draw X cards. Activate only as a sorcery.
//	   {2}{R}{R}{G}{G}: Transform Bruce Banner. Activate only as a
//	   sorcery."
//
//	The Incredible Hulk {2}{R}{R}{G}{G}, Legendary Creature — Gamma
//	Berserker Hero, 8/8
//	  "Reach, trample
//	   Enrage — Whenever The Incredible Hulk is dealt damage, put a +1/+1
//	   counter on him. If he's attacking, untap him and there is an
//	   additional combat phase after this phase."
//
// {X}{X} is twice X (the Folio of Fancies ruling) and X is read back as
// the ability resolves. The transform is the permanent turning over in
// place (Transform, ADR 0079). Enrage is one trigger per damage event, not
// per source (CR 120.3): it waits for the damage to finish (AtBatchEnd,
// Strong's shape), so Hulk blocked by three creatures triggers once. The
// counter goes on first, then, only if he is attacking when the trigger
// resolves, he untaps and the turn gets an additional combat phase after
// the current phase (CR 500.8), with no main phase between.
//
// No simplification.
func init() {
	const oracle = "b6c0ee14-2028-4e8b-9acd-ff78821bca9b"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Bruce Banner",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:        "{X}{X}, {T}: Draw X cards. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{X}{X}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
			},
		}, {
			Label:        "{2}{R}{R}{G}{G}: Transform Bruce Banner. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{R}{R}{G}{G}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "The Incredible Hulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "The Incredible Hulk — Enrage: put a +1/+1 counter on him; if he's attacking, untap him and there is an additional combat phase",
			AtBatchEnd: ThatMuchDamage(),
			Effect:     incredibleHulkEnrage,
		}},
	})
}

// incredibleHulkEnrage is the Enrage body: the counter, then, if he is
// attacking, the untap and the extra combat phase.
func incredibleHulkEnrage(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || self.AttackingTarget == uuid.Nil {
		return nil
	}
	if err := (UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
