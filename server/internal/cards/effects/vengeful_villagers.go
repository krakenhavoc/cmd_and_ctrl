package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengeful Villagers — Creature — Human Citizen {3}{W}, 3/3 (Avatar: The
// Last Airbender, #4459):
//
//	"Whenever this creature attacks, choose target creature an opponent
//	 controls. Tap it, then you may sacrifice an artifact or creature. If
//	 you do, put a stun counter on the chosen creature. (If a permanent
//	 with a stun counter would become untapped, remove one from it
//	 instead.)"
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b): a creature that left, or changed hands to you,
// takes the whole trigger with it. It is tapped first (CR 701.26a); then
// you pick up to one artifact or creature you control to sacrifice — the
// "you may" is a pick of zero or one, so declining is picking none, and a
// permanent that can't be sacrificed is not offered. The Villagers
// themselves are an eligible creature. Only once the sacrifice has
// happened (CR 701.21a) does the stun counter go on the chosen creature
// (CR 122.1d), and only while it is still the creature that was targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad6442c2-7aa2-4c24-8133-f679a3499bde",
		Name:         "Vengeful Villagers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Vengeful Villagers — tap target creature an opponent controls; you may sacrifice an artifact or creature to stun it",
					vengefulVillagersTapThenOffer),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// vengefulVillagersTapThenOffer taps the target, then offers the
// sacrifice.
func vengefulVillagersTapThenOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	var candidates []uuid.UUID
	for _, id := range permanentsControlledByMatching(g, item.Controller, Or(Artifact(), Creature())) {
		if !g.CantBeSacrificedForEffect(id) {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Vengeful Villagers — you may sacrifice an artifact or creature; if you do, the chosen creature gets a stun counter",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    vengefulVillagersSacrifice,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// vengefulVillagersSacrifice is the pick: none is the "no", one is
// sacrificed and the stun follows once it has gone. An on-disk key.
var vengefulVillagersSacrifice = askedPick("choose-cards/vengeful-villagers-sacrifice", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, ctx.ItemTail(vengefulVillagersStun))
})

// vengefulVillagersStun is "if you do, put a stun counter on the chosen
// creature": told what was sacrificed, it stuns the target if it is still
// the legal target it was. An on-disk key.
var vengefulVillagersStun = CardTail("then/vengeful-villagers-stun", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return AddCounter{Target: target, Kind: game.CounterStun, N: 1}.Apply(ctx)
})
