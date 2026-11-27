package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b565_helpers.go — shared pieces of Coldsnap batch 565 (#5597, S82).

// S82OldB565IceCounter is the ice counter's kind (Coldsnap, CR 122.1).
const S82OldB565IceCounter = "ice"

// S82OldB565BlocksAFlyer is "whenever this creature blocks a creature with
// flying" (CR 509.3b): its own block, against an attacker that has flying
// as the block is declared.
func S82OldB565BlocksAFlyer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !B28SelfBlocked(ev, source) {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.Target)
	return ok && game.HasKeyword(&attacker, "flying")
}

// S82OldB565PreventDamageDealtBy is a standing CR 615 prevention of damage
// dealt by the permanent that has it, to any recipient: "prevent all
// [combat] damage it would deal". `only` narrows it further (nil: none), read
// as the damage would be dealt; `combatOnly` limits it to combat damage.
func S82OldB565PreventDamageDealtBy(label string, combatOnly bool, only func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource != src.InstanceID {
				return false
			}
			if combatOnly && !ev.IsCombatDamage {
				return false
			}
			return only == nil || only(ev, g, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// S82OldB565TapAndIceTheTarget is "Tap target creature and put an ice counter
// on it" as an ability's resolution. A target gone by then is no target
// (CR 608.2b).
func S82OldB565TapAndIceTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return AddCounter{Target: t.ID, Kind: S82OldB565IceCounter, N: 1}.Apply(ctx)
	}
	return nil
}

// S82OldB565IceCounterOnTheTarget is "Put an ice counter on target
// permanent".
func S82OldB565IceCounterOnTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: S82OldB565IceCounter, N: 1}.Apply(ctx)
		}
	}
	return nil
}

// S82OldB565SnowPermanentsOnTheBattlefield counts every snow permanent,
// whoever controls it, read through the layers so far (an ice-counter
// permanent has been made snow by layer 4 before layer 7 asks).
func S82OldB565SnowPermanentsOnTheBattlefield(g *game.Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].HasSupertype("snow") {
			n++
		}
	}
	return n
}
