package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CantBeBlockedThisTurnByPlayer registers "<target creature> can't be
// blocked by creatures that player controls this turn" (The Black Gate,
// #2172, CR 509.1b). It is a cantBeBlockedByPlayer ScopedEffect record
// pinned to Target at resolution (CR 611.2c) and swept at the turn's
// end; the block validator and the enumerator read it through the same
// block-rule walk as the other kinds, so a barred blocker is refused and
// never offered.
type CantBeBlockedThisTurnByPlayer struct {
	// Target pins the rule to one permanent.
	Target uuid.UUID
	// Player is the seat whose creatures may not block it. A zero
	// Player registers nothing (the choice could not be made).
	Player uuid.UUID
	// Label names the effect in logs and the continuation census.
	Label string
}

// Apply registers the record. Caller is inside the resolution frame
// (holds g.mu write).
func (r CantBeBlockedThisTurnByPlayer) Apply(ctx *Context) error {
	if r.Player == uuid.Nil {
		return nil
	}
	name := "that player"
	if p := ctx.Game.PlayerByIDForEffect(r.Player); p != nil && p.Name != "" {
		name = p.Name
	}
	text := "creatures " + name + " controls"
	return ScopedEffectFor{
		Target:   r.Target,
		Mods:     []game.Mod{game.CantBeBlockedByPlayerMod(r.Player, text)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(r.Label, "can't be blocked this turn by "+text),
	}.Apply(ctx)
}

// CantBeBlockedThisTurnByPower registers "<target creature> can't be
// blocked by creatures with power N or less this turn" (Rhonas's
// Stalwart, #2600, CR 509.1b). It is a cantBeBlockedByPower
// ScopedEffect record pinned to Target at resolution (CR 611.2c) and
// swept at the turn's end. The blocker's power is read live as blockers
// are declared, so the validator and the enumerator both see a pump or
// shrink made after the effect.
type CantBeBlockedThisTurnByPower struct {
	// Target pins the rule to one permanent.
	Target uuid.UUID
	// AtMost is N: a blocker whose power is N or less is barred.
	AtMost int
	// Label names the effect in logs and the continuation census.
	Label string
}

// Apply registers the record. Caller is inside the resolution frame
// (holds g.mu write).
func (r CantBeBlockedThisTurnByPower) Apply(ctx *Context) error {
	return ScopedEffectFor{
		Target:   r.Target,
		Mods:     []game.Mod{cantBeBlockedByPowerMod(r.AtMost)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(r.Label, "can't be blocked this turn by creatures with power "+strconv.Itoa(r.AtMost)+" or less"),
	}.Apply(ctx)
}

// cantBeBlockedByPowerMod is the mod CantBeBlockedThisTurnByPower
// registers, shared with the cards that register it beside another mod
// in one record.
func cantBeBlockedByPowerMod(n int) game.Mod {
	return game.CantBeBlockedByPowerMod(n, "creatures with power "+strconv.Itoa(n)+" or less")
}
