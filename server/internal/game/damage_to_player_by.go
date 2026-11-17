package game

import "github.com/google/uuid"

// damage_to_player_by.go — "it has dealt N or more damage to that player
// this turn" (#5674).
//
// TurnTally.DamagedPlayers lists every damage event dealt to a player
// this turn: the source OBJECT (instance ID and the ObjectEpoch it had as
// the damage was dealt), the player, and the amount actually dealt after
// replacement and prevention. Combat and noncombat damage alike are
// recorded, one entry per event, and the turn boundary clears it with the
// rest of the tally. It is the player-side sibling of DamagedObjects
// (dealt_damage_by.go), which records damage to permanents: DamageDealers
// names which creatures hit which player but is deduplicated and carries
// no amount, and NoncombatDamage has amounts for noncombat damage only.
//
// The amount is the damage, not the life it cost (CR 120.3): damage from
// a source with infect, or to a player whose life total can't change, is
// still damage dealt.
//
// Object identity is the point (CR 400.7): a source that left and came
// back is a new object that has dealt nothing. The record outlives the
// source, so an ability of a creature that died to the same combat still
// reads what that object dealt (CR 608.2h).

// PlayerDamageRecord is one damage event dealt to a player (#5674): the
// source object as it was when it dealt the damage, the damaged player
// and the amount dealt.
type PlayerDamageRecord struct {
	Source ObjectRef `json:"source"`
	Victim uuid.UUID `json:"victim"`
	Amount int       `json:"amount"`
}

// recordDamagedPlayerLocked notes the damage event `ev`, dealt to a
// player. A source the engine cannot find anywhere is not recorded,
// which errs weaker: nothing can count it. Caller must hold g.mu.
func (g *Game) recordDamagedPlayerLocked(ev Event) {
	if ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil {
		return
	}
	src := g.findCardByIDLocked(ev.Source)
	if src == nil {
		return
	}
	g.TurnTally.DamagedPlayers = append(g.TurnTally.DamagedPlayers, PlayerDamageRecord{
		Source: ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch},
		Victim: ev.Target,
		Amount: ev.Amount,
	})
}

// DamageDealtToPlayerThisTurnByObjectForEffect is the total damage the
// object `source` has dealt to the player `victim` this turn, combat and
// noncombat together: Vessel of the All-Consuming's "if it has dealt 10
// or more damage to that player this turn". A ref naming a different
// object of the same card, one that left and came back, counts nothing
// (CR 400.7), and the object does not have to exist any more
// (CR 608.2h).
//
// The tally listener runs ahead of the trigger harvester, so a trigger
// on the damage event that asks already counts that event.
//
// Caller must hold g.mu.
func (g *Game) DamageDealtToPlayerThisTurnByObjectForEffect(source ObjectRef, victim uuid.UUID) int {
	if source.ID == uuid.Nil || victim == uuid.Nil {
		return 0
	}
	total := 0
	for _, rec := range g.TurnTally.DamagedPlayers {
		if rec.Source == source && rec.Victim == victim {
			total += rec.Amount
		}
	}
	return total
}
