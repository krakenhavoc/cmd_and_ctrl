package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB610LandWentToGraveyard is "a land is put into a graveyard from the
// battlefield" under any player's control (Dingus Egg): the exit read
// post-move, as a land as it last existed on the battlefield (CR 603.10a).
func OldB610LandWentToGraveyard(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && LeftAsType(ev, c, "land")
}

// OldB610YourCreatureDealtCombatDamage is "a creature you control deals combat
// damage" (Noble Purpose): any recipient, a positive amount actually dealt. The
// creature is read live, since combat damage lands before the state-based
// sweep (CombatDamageToPlayerBy).
func OldB610YourCreatureDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature() && src.Controller == source.Controller
}

// OldB610ThisDealsOneToTheTarget is "it deals 1 damage to target …" for a
// trigger's own source (Elite Javelineer): the damage is dealt by the object
// that triggered, by last-known information if it has left (CR 608.2h), to the
// first target that is still legal (CR 608.2b).
func OldB610ThisDealsOneToTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{SourceObject: &ref, Target: t.ID, Amount: 1}.Apply(ctx)
	}
	return nil
}
