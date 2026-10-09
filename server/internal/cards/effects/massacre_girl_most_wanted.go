package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Massacre Girl, Most Wanted — Legendary Creature — Human Assassin
// {4}{B}, 4/4 (Reality Fracture):
//
//		"Whenever another creature or planeswalker you control dies,
//		 Massacre Girl deals 1 damage to target opponent and you gain 1
//		 life.
//		 Whenever an opponent is dealt noncombat damage, put a +1/+1 counter
//		 on Massacre Girl."
//
//	  - The dies trigger reads what the permanent last was (CR 603.10a), so
//	    a planeswalker and an animated manland both count, and "you control"
//	    is who controlled it as it left.
//	  - The damage and the life are one ability: if its target opponent is
//	    gone by resolution, neither happens.
//	  - The counter trigger fires per damage event to a player other than
//	    Massacre Girl's controller, combat damage excluded.
func init() {
	Register(Spec{
		OracleID:     "81f35914-73bf-429f-ba20-32e23ac30159",
		Name:         "Massacre Girl, Most Wanted",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventLTB, anotherCreatureOrPlaneswalkerYouControlDied,
					"Massacre Girl, Most Wanted — 1 damage to target opponent and you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						victim := uuid.Nil
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								victim = t.ID
								break
							}
						}
						if victim == uuid.Nil {
							return nil
						}
						if err := (DealDamage{Source: item.SourceCardID, Target: victim, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
						return GainLife{Amount: 1}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent()),
			),
			On(game.EventDealDamage, anOpponentWasDealtNoncombatDamage,
				"Massacre Girl, Most Wanted — put a +1/+1 counter on Massacre Girl",
				plusOneCountersOnThis(1)),
		},
	})
}

// anotherCreatureOrPlaneswalkerYouControlDied is "another creature or
// planeswalker you control dies": a graveyard arrival of something that
// was a creature or planeswalker as it last existed, under the source's
// controller.
func anotherCreatureOrPlaneswalkerYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	if !leftAsType(ev, c, "creature") && !leftAsType(ev, c, "planeswalker") {
		return false
	}
	return leftUnderControlOf(ev, c) == source.Controller
}

// anOpponentWasDealtNoncombatDamage is "whenever an opponent is dealt
// noncombat damage": a damage event that is not combat damage, to a
// player other than the source's controller.
func anOpponentWasDealtNoncombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 || ev.Target == uuid.Nil || ev.Target == source.Controller {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}
