package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bomb Squad — Creature — Dwarf {3}{R}, 1/1:
//
//	"{T}: Put a fuse counter on target creature.
//	 At the beginning of your upkeep, put a fuse counter on each creature
//	 with a fuse counter on it.
//	 Whenever a creature has four or more fuse counters on it, remove all
//	 fuse counters from it and destroy it. That creature deals 4 damage
//	 to its controller."
//
// #1858. The last ability is a CR 603.8 state trigger about each
// creature: "whenever" in the printed text, but a state, and a state of
// one creature. It triggers once for each creature that has four or more
// fuse counters, whoever controls it (Bomb Squad included), and not
// again for that creature while its instance waits or is on the stack.
// Two Bomb Squads each trigger for the same creature, as the card's
// ruling says, and both deal their 4 damage although the first destroys
// it.
//
// On resolution the counters come off and the creature is destroyed if
// it is still the object that triggered the ability (CR 400.7). The
// creature then deals 4 damage to its controller either way: a creature
// that regenerated still deals it (the card's ruling), and one that left
// in response deals it by last-known information, with the lifelink,
// infect or wither it had (CR 608.2h). The trigger has no intervening
// "if", so it resolves even if counters came off in response.
//
// No simplification.
func init() {
	const fuse = "fuse"
	Register(Spec{
		OracleID:     "f2ecb354-8f79-4c39-989e-7aa37ec75154",
		Name:         "Bomb Squad",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a fuse counter on target creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  putACounterOnTheTarget(fuse),
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bomb Squad — put a fuse counter on each creature with a fuse counter on it",
				putACounterOnEachCreatureWith(fuse)),
			WheneverACreatureHasAtLeast(fuse, 4, "Bomb Squad — remove the fuse counters and destroy that creature",
				bombSquadDetonate),
		},
	})
}

// bombSquadDetonate is the state trigger's effect: "remove all fuse
// counters from it and destroy it. That creature deals 4 damage to its
// controller."
func bombSquadDetonate(g *game.Game, item *game.StackItem) error {
	const fuse = "fuse"
	ctx := NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	ref := obj.Ref()
	controller := obj.Controller
	if creature, ok := ctx.TriggeringPermanent(); ok {
		controller = creature.Controller
		if !creature.Left {
			if n := creature.Counters[fuse]; n > 0 {
				if err := (AddCounter{Target: ref.ID, Kind: fuse, N: -n}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return DealDamage{SourceObject: &ref, Target: controller, Amount: 4}.Apply(ctx)
}
