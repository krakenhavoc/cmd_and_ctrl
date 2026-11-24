package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of Truth and Justice — Artifact — Equipment {3}:
//
//	"Equipped creature gets +2/+2 and has protection from white and from blue.
//	 Whenever equipped creature deals combat damage to a player, put a +1/+1
//	 counter on a creature you control, then proliferate.
//	 Equip {2}"
//
// The Swords' shared shape: the pump and ADR 0072's layer-6 protection grant
// on the equipped creature, and a combat-damage trigger matched through the
// attachment. The counter does not target: the creature is chosen as the
// trigger resolves (Choking Miasma's prompt), so a creature with hexproof or
// protection can be chosen, and the proliferate waits for the counter to land
// so it can proliferate that counter. With no creature to choose the counter is
// skipped and the proliferate still happens. Each proliferate chooses any number
// of permanents and players (CR 701.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a8baaf9-e21f-41a0-9f15-80764f6e5e68",
		Name:         "Sword of Truth and Justice",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("protection from white", "protection from blue"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Sword of Truth and Justice — put a +1/+1 counter on a creature you control, then proliferate",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(effects.CreaturesControlledByPlayer(g, ctx.Controller())) == 0 {
					return effects.Proliferate{}.Apply(ctx)
				}
				return effects.ChoosePermanents{
					Question:   "Sword of Truth and Justice — put a +1/+1 counter on a creature you control",
					Candidates: effects.HumiliateCreatures,
					Keyed:      swordOfTruthAndJusticeChosen,
				}.Apply(ctx)
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// swordOfTruthAndJusticeChosen puts the counter on the creature chosen, then
// proliferates. An on-disk key.
var swordOfTruthAndJusticeChosen = effects.RegisterPermanentsPick("prompt-run/sword-of-truth-and-justice-counter", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.Proliferate{}.Apply(ctx)
	},
})
