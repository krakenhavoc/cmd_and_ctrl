package game

import "github.com/google/uuid"

// speed.go — Aetherdrift's speed: start your engines! (CR 702.179) and
// max speed (CR 702.178). ADR 0136, #2122.
//
// Speed belongs to a PLAYER, not to a permanent: a player has none
// until a rule or effect sets it (CR 702.179b), and nothing in the
// rules ever lowers it, so a player keeps it after the permanent that
// started it has gone. Four rules make the whole mechanic, each with
// one home below:
//
//   - CR 702.179a / 704.5aa: a player with no speed who controls a
//     permanent with start your engines! gets speed 1. A state-based
//     action: startYourEnginesSBALocked, called from
//     stateBasedActionsLocked.
//   - CR 702.179d: a player who has speed has an inherent triggered
//     ability with no source — "Whenever one or more opponents lose
//     life during your turn, if your speed is less than 4, increase
//     your speed by 1. This ability triggers only once each turn." It
//     uses the stack. speedTriggers, a Listener registered beside the
//     monarch's, for the monarch's reason: there is no card for the
//     harvester to find the ability on.
//   - CR 702.179d / 702.179e: speed stops at 4, and a player has max
//     speed while it is 4. setSpeedLocked is the one write and refuses
//     anything outside 0..4.
//   - CR 702.178a: "Max speed — [ability]" is "as long as your speed is
//     4, this object has [ability]". HasMaxSpeed is the one read;
//     the card-side constructors in effects/speed.go put it in each
//     ability's own predicate (ADR 0136 §5 says why that is not an
//     ADR 0071 designation).
//
// Speed is not a counter (nothing proliferates, doubles or removes it),
// so it is a plain int on Player and not an entry in Player.Counters.

// MaxSpeed is the highest speed a player can have (CR 702.179d's "if
// your speed is less than 4"), and the speed at which a player "has
// max speed" (CR 702.179e).
const MaxSpeed = 4

// KeywordStartYourEngines is the canonical token for CR 702.179's
// keyword: the deck importer's lowercase of Scryfall's "Start your
// engines!". Named because the state-based action, the catalog and the
// registry all read it, and a typo would leave a deck that never gets
// speed.
const KeywordStartYourEngines = "start your engines!"

// EventSpeedChanged — a player's speed changed. Actor is the player,
// Amount the NEW speed (1 to MaxSpeed). Emitted by setSpeedLocked only,
// and only on a real change. A layer input (layerVersionBump), because
// a "Max speed —" static reads it. ADR 0136.
const EventSpeedChanged EventKind = "speed_changed"

// SpeedOf is player's speed: 0 while they have none, which is also what
// an effect that refers to speed reads then (CR 702.179f). Caller must hold
// g.mu. Reads the Player and nothing else, so a static's AppliesTo may
// call it from inside the layer pass.
func (g *Game) SpeedOf(player uuid.UUID) int {
	p := g.playerByIDLocked(player)
	if p == nil {
		return 0
	}
	return p.Speed
}

// HasMaxSpeed reports whether player has max speed (CR 702.179e): a
// speed of MaxSpeed. Caller must hold g.mu.
func (g *Game) HasMaxSpeed(player uuid.UUID) bool {
	return g.SpeedOf(player) >= MaxSpeed
}

// setSpeedLocked is the one write to Player.Speed. A value outside
// 0..MaxSpeed is refused, a value equal to the current one is nothing,
// and every real change emits EventSpeedChanged.
//
// Caller must hold g.mu in write mode.
func (g *Game) setSpeedLocked(player uuid.UUID, speed int) {
	if speed < 0 || speed > MaxSpeed {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || p.Speed == speed {
		return
	}
	p.Speed = speed
	g.EmitEvent(Event{
		Kind:   EventSpeedChanged,
		Actor:  player,
		Amount: speed,
	})
}

// SetSpeedForTest sets player's speed through the one write, event and
// all, for a test outside the package that needs a player at a given
// speed without playing the turns to get there. Takes the write lock.
func (g *Game) SetSpeedForTest(player uuid.UUID, speed int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.setSpeedLocked(player, speed)
}

// increaseSpeedLocked is "increase your speed by 1": nothing for a
// player already at max speed, and speed 1 for a player with none
// (CR 702.179c: "their speed becomes that value"). The inherent trigger
// only exists for a player with speed, so the second case is the verb
// a future "increase your speed" card would reach.
//
// Caller must hold g.mu in write mode.
func (g *Game) increaseSpeedLocked(player uuid.UUID) {
	cur := g.SpeedOf(player)
	if cur >= MaxSpeed {
		return
	}
	g.setSpeedLocked(player, cur+1)
}

// startYourEnginesSBALocked is CR 702.179a / 704.5aa: each player still in the
// game who has no speed and controls a permanent with start your
// engines! gets speed 1. Reports whether it gave anyone speed.
//
// HasKeyword reads the effective abilities, so a granted instance
// counts. A phased-out permanent is not in the battlefield slice and
// does not (CR 702.26b). A player who already has speed is never
// touched, so one whose last speed permanent has left keeps it
// (CR 702.179b: only a rule or effect sets a speed).
//
// Caller must hold g.mu in write mode.
func (g *Game) startYourEnginesSBALocked() bool {
	fired := false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Speed > 0 {
			continue
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != p.ID || !HasKeyword(c, KeywordStartYourEngines) {
				continue
			}
			g.setSpeedLocked(p.ID, 1)
			fired = true
			break
		}
	}
	return fired
}

// speedTriggers is the Listener half of CR 702.179d.
type speedTriggers struct{}

// OnEvent runs under g.mu in write mode from notifyListenersLocked, so
// everything it calls is a *Locked helper.
func (speedTriggers) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventChangeLife:
		if ev.Amount < 0 {
			g.speedTriggerLocked(ev.Target)
		}
	case EventDealDamage:
		// #2105: the life the damage COST. Infect damage, and damage
		// to a player whose life total can't change, loses no life.
		// A creature or planeswalker target has no player entry and
		// is turned away by speedTriggerLocked.
		if ev.DamageLifeLoss() > 0 {
			g.speedTriggerLocked(ev.Target)
		}
	}
}

// speedTriggerLocked is CR 702.179d firing for a loss of life by
// `loser`: the ACTIVE player's ability, because it is "during your
// turn", and only when the loser is one of their opponents. Damage
// that costs life is a loss (CR 120.3a), and so is life paid
// (CR 119.4).
//
// Every gate is read as the ability triggers: the active player is
// still in the game and has speed, the speed is below MaxSpeed (the
// intervening if, CR 603.4, read again by the body), and the ability
// has not triggered this turn. The once-per-turn flag is set as the
// item is queued, so a second loss in the same turn — another
// creature's combat damage, a second drain, a second opponent in the
// same "each opponent loses 1 life" — finds it set, which is also what
// "one or more opponents" means for a simultaneous loss.
//
// Caller must hold g.mu in write mode.
func (g *Game) speedTriggerLocked(loser uuid.UUID) {
	if loser == uuid.Nil {
		return
	}
	active := g.activePlayerIDLocked()
	if active == uuid.Nil || active == loser {
		return
	}
	if g.playerByIDLocked(loser) == nil {
		return
	}
	p := g.playerByIDLocked(active)
	if p == nil || p.Eliminated || p.Speed <= 0 || p.Speed >= MaxSpeed {
		return
	}
	if g.TurnTally.Players[active].SpeedTriggered {
		return
	}
	g.bumpPlayerTally(active, func(t *PlayerTurnTally) { t.SpeedTriggered = true })
	params := EffectParams{Player: active}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind: StackItemTriggered,
		// CR 702.179d: the ability has no source and is controlled by
		// the player whose speed it raises.
		Controller: active,
		Owner:      active,
		Label:      "Speed — increase " + p.Name + "'s speed",
		Body:       speedIncreaseBody.Key(),
		Params:     params,
		Effect:     bodyEffect(speedIncreaseBody.Key(), params),
	})
}

// speedIncreaseBody is "speed/increase" (ADR 0136 §3): Params.Player's
// speed increases by 1 if it is still below MaxSpeed — the second half
// of CR 603.4's intervening if, which increaseSpeedLocked's own cap
// already is (the trigger only exists for a player with speed, so the
// CR 702.179c branch never runs from here).
var speedIncreaseBody = DelayedBody("speed/increase", func(g *Game, _ *StackItem, p EffectParams) error {
	g.increaseSpeedLocked(p.Player)
	return nil
})
