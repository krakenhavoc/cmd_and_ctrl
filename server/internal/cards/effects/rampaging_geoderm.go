package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampaging Geoderm — Creature — Dinosaur Beast {2}{R}{G}, 3/3 (March of
// the Machine, #5649):
//
//	"Trample, haste
//	 Whenever you attack, target attacking creature gets +1/+1 until end
//	 of turn. If it's attacking a battle, put a +1/+1 counter on it
//	 instead."
//
// One trigger per attack declaration of yours, however many creatures
// attack (CR 508.3d), aimed at any attacking creature. What the target is
// attacking is read as the ability resolves: a creature attacking a battle
// then gets the counter and no pump, any other gets the pump and no
// counter. A creature whose battle was removed from combat is attacking
// nothing (CR 506.4c) and gets the pump. A target that has stopped
// attacking is illegal and gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e536e0d2-7117-4ca8-bdb4-e7dd380113c6",
		Name:            "Rampaging Geoderm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Rampaging Geoderm — target attacking creature gets +1/+1 until end of turn, or a +1/+1 counter instead if it's attacking a battle",
					rampagingGeodermPump),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

// rampagingGeodermPump puts a +1/+1 counter on the still-legal target if
// it is attacking a battle, and gives it +1/+1 until end of turn if not.
func rampagingGeodermPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if isAttackingABattle(g, c) {
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Rampaging Geoderm — +1/+1"}.Apply(ctx)
}
