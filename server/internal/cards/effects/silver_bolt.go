package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Bolt — Artifact {1} (Innistrad: Midnight Hunt):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 3 damage to target
//	 creature. If a Werewolf is dealt damage this way, destroy it."
//
// The damage is dealt as the ability resolves and the Werewolf check is made
// on what was actually dealt (The Black Arrow's keyed tail), so damage that
// is prevented does not destroy. "Werewolf" is a word of this card's text
// (CR 612.2). The artifact is sacrificed as a cost, so the ability is the
// source of the damage by last-known information (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b29a0686-2246-47a9-974d-4425e8a5546c",
		Name:         "Silver Bolt",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature. If a Werewolf is dealt damage this way, destroy it.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  silverBoltShot,
		}},
	})
}

func silverBoltShot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return g.DealDamageToCreatureKeyedForEffect(item.SourceCardID, targets[0].ID, 3, ctx.ItemTail(silverBoltDestroyWerewolf))
}

// silverBoltDestroyWerewolf is "if a Werewolf is dealt damage this way,
// destroy it", told the permanent and the damage it was dealt. An on-disk key.
var silverBoltDestroyWerewolf = CardTail("then/silver-bolt-destroy-werewolf", func(ctx *Context, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !c.HasSubtype(ctx.TextSubtype("Werewolf")) {
		return nil
	}
	return nothingIfGone(ctx.Game.DestroyPermanentForEffect(o.IDs[0]))
})
