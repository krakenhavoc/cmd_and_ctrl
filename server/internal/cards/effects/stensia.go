package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stensia — Plane — Innistrad (Commander Masters, #4870):
//
//	"Whenever a creature deals damage to one or more players for the
//	 first time each turn, put a +1/+1 counter on it.
//	 Whenever chaos ensues, each creature you control gains "{T}: This
//	 creature deals 1 damage to target player or planeswalker" until end
//	 of turn."
//
// A plane (ADR 0216 §6). The first ability triggers on any creature's damage
// to a player, combat or not, the first time that creature object deals
// damage to a player this turn (game.TurnTally.DamageEventsToPlayers, a
// per-object count, so a creature that left and came back is a new one, CR
// 400.7); the counter goes on the creature that dealt it, if it is still on
// the battlefield, and a creature that has already had one this turn gets
// no second one (the ruling). The chaos ability (CR 311.7) grants a
// tap-for-damage ability to each creature the planar controller controls as
// it resolves (CR 611.2c) until end of turn; a creature that came under
// that control this turn can't use it until it has been controlled
// continuously since the turn began (CR 302.6), like any {T} ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38bdd034-0375-403c-97ca-6f93a5e12e54",
		Name:         "Stensia",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: stensiaPingGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
				Cost:    TapCost(),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to target player or planeswalker.",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, aCreatureDealtDamageToAPlayerForTheFirstTimeThisTurn,
				"Stensia — put a +1/+1 counter on that creature", stensiaCounter),
			WheneverChaosEnsues("Stensia — each creature you control gains \"{T}: This creature deals 1 damage to target player or planeswalker\" until end of turn",
				Do(GrantAbilitiesFor{
					Match: And(Creature(), YouControl()),
					Keys:  []string{stensiaPingGrant},
					Label: "Stensia — {T}: 1 damage to target player or planeswalker",
				})),
		},
	})
}

// stensiaPingGrant is the granted bundle's key. An on-disk key.
const stensiaPingGrant = "stensia/tap-for-1-damage"

// aCreatureDealtDamageToAPlayerForTheFirstTimeThisTurn is Stensia's
// condition: damage to a player from a creature, the first damage event
// that creature object has dealt to a player this turn. The turn tally
// counts ahead of the harvester, so the event itself is the count of 1.
func aCreatureDealtDamageToAPlayerForTheFirstTimeThisTurn(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.IsCreature() && g.DamageEventsToPlayersThisTurn(ev.Source) == 1
}

// stensiaCounter puts the counter on the creature that dealt the damage,
// unless it has already had one from this plane this turn.
func stensiaCounter(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	dealer := item.Trigger.Event.Source
	if !onBattlefield(g, dealer) {
		return nil
	}
	return AddCounter{Target: dealer, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
