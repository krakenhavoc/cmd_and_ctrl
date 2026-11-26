package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB430SwampEnteredUnderYourControl is "a Swamp you control enters"
// (Veilborn Ghoul): a land with the Swamp subtype entering under the source
// owner's control, read from the graveyard as a Bloodghast-shaped trigger does.
func S82OldB430SwampEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsLand() && c.HasSubtype(source.TextSubtype("Swamp"))
}

// S82OldB430SaprolingsForGreatestPower is Fungal Sprouting: "Create X 1/1
// green Saproling creature tokens, where X is the greatest power among
// creatures you control", read as the spell resolves (CR 608.2h).
func S82OldB430SaprolingsForGreatestPower(_ *game.StackItem, ctx *Context) error {
	g := ctx.Game
	g.RecomputeLayersIfStaleLocked()
	n := B42GreatestPowerControlledBy(g, ctx.Controller())
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
}

// S82OldB430ExileAllFromHandsAndGraveyards is Worldfire's second sentence:
// "Exile all cards from all hands and graveyards", one simultaneous batch. The
// ids are snapshotted before anything moves.
func S82OldB430ExileAllFromHandsAndGraveyards(ctx *Context) error {
	var ids []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil {
			continue
		}
		if p.Hand != nil {
			for _, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
		}
		if p.Graveyard != nil {
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	ctx.Game.ExileCardsForEffect(ids)
	return nil
}

// S82OldB430EachPlayerLifeBecomesOne is Worldfire's third sentence: each
// player's life total becomes 1 (CR 119.5), a gain or a loss of the difference,
// for every player still in the game in turn order.
func S82OldB430EachPlayerLifeBecomesOne(ctx *Context) error {
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := B31LifeBecomes(ctx, p.ID, 1); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB430SixDamageToEachTarget is Chandra, the Firebrand's −6: she deals 6
// damage to each of up to six targets, as one damage event, to each target
// still legal.
func S82OldB430SixDamageToEachTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	return g.DamageInstanceForEffect(func() error {
		for _, t := range targets {
			if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 6}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
