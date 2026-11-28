package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB841SnowSwampsYouControl is how many snow Swamps `player` controls, read
// through the layered view ("if you control a snow Swamp", Gangrenous Zombies;
// "the number of snow Swamps you control", Withering Wisps).
func S82OldB841SnowSwampsYouControl(g *game.Game, player uuid.UUID) int {
	return len(PermanentsControlledByMatching(g, player, And(HasSubtype("Swamp"), Snow())))
}

// S82OldB841MonsoonTapIslands is Monsoon's trigger as it resolves: the player whose end
// step it is taps every untapped Island they control, then takes that many damage from
// the Monsoon. Only lands that actually became tapped count ("the number of Islands
// tapped this way"); a Monsoon gone by then deals its damage as it last existed.
func S82OldB841MonsoonTapIslands(g *game.Game, item *game.StackItem) error {
	who := item.Trigger.Event.Actor
	if who == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	tapped := 0
	for _, id := range PermanentsControlledByMatching(g, who, And(HasSubtype("Island"), Land())) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Tapped {
			continue
		}
		if err := g.TapTargetForEffect(id); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: who, Amount: tapped}.Apply(ctx)
}

// S82OldB841GangrenousZombiesDamage is "This creature deals 1 damage to each creature and
// each player. If you control a snow Swamp, this creature deals 2 damage to each
// creature and each player instead." The snow Swamp is read as the ability resolves.
func S82OldB841GangrenousZombiesDamage(g *game.Game, item *game.StackItem) error {
	n := 1
	if S82OldB841SnowSwampsYouControl(g, item.Controller) > 0 {
		n = 2
	}
	return B23DamageEachCreatureAndEachPlayer(NewContext(g, item), n)
}

// S82OldB841HalfSacrificedPowerDamage is "This creature deals damage to any target equal
// to half the sacrificed creature's power, rounded down" (Freyalise Supplicant): the
// sacrificed creature's last-known power, none for a power below 2 (CR 107.1b).
func S82OldB841HalfSacrificedPowerDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := ctx.SacrificedPower() / 2
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// S82OldB841DividedXPlusOne is Meteor Shower's resolution: the announced division of X plus 1.
func S82OldB841DividedXPlusOne(_ *game.StackItem, ctx *Context) error { return DealDividedDamage(ctx) }
