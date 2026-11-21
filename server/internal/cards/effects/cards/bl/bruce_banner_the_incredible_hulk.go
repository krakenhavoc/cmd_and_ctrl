package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Bruce Banner",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}{X}, {T}: Draw X cards. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{X}{X}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
			},
		}, {
			Label:        "{2}{R}{R}{G}{G}: Transform Bruce Banner. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{2}{R}{R}{G}{G}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "The Incredible Hulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "The Incredible Hulk — Enrage: put a +1/+1 counter on him; if he's attacking, untap him and there is an additional combat phase",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     incredibleHulkEnrage,
		}},
	})
}

// incredibleHulkEnrage is the Enrage body: the counter, then, if he is
// attacking, the untap and the extra combat phase.
func incredibleHulkEnrage(g *game.Game, item *game.StackItem) error {
	if !effects.B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || self.AttackingTarget == uuid.Nil {
		return nil
	}
	if err := (effects.UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return effects.ExtraCombatAfterThisPhase().Apply(ctx)
}
