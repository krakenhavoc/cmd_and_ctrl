package game

import (
	"fmt"

	"github.com/google/uuid"
)

// player_mana_ability.go — #4885: a mana ability a PLAYER has for a
// duration, bound to no card.
//
//	Channel  "Until end of turn, any time you could activate a mana
//	          ability, you may pay 1 life. If you do, add {C}."
//
// Every other mana ability in the engine is a row on a card
// (ActivateManaAbility takes a card id and resolves an index against
// the card's rows). Channel's is not: the spell is in a graveyard a
// moment after it resolves, and the permission outlives it. So it is
// stored where every other granted player ability is, on
// Player.Statics, as one more payload-told kind of PlayerStatic
// (player_statics.go): plain data, a CR 611.2 duration swept by
// sweepPlayerStaticsLocked at cleanup, cloned for undo and mirrored
// into the snapshot with the rest of the slice. No closure, so nothing
// for the restorability census to count.
//
// WHERE IT CAN BE USED. "Any time you could activate a mana ability"
// is exactly the set of moments the wire's activate_mana_ability verb
// is accepted, so the activation rides that verb (its `player_ability`
// field names the grant) and inherits every gate that already admits a
// mana activation: holding priority, answering a "pay {N}" prompt, a
// controlled seat's playing stage (actions.controlled_player.go). As
// with a card's mana ability, mana floated this way before a cast pays
// it (CR 601.2g, CR 605.3a).
//
// WHAT THE PLANNER DOES WITH IT: nothing. The auto-tapper never spends
// life on the player's behalf (a mana ability's life cost "stays out of
// the plan entirely", docs/adding-cards.md), and it does not gather
// this source, so a cast the player could only afford with it is
// refused by the strict gate until the player floats the mana
// themselves. That is the weaker direction and never a wrong plan.

// PlayerManaAbility is the payload of a granted player mana ability:
// "pay LifeCost life: add Produced". Plain data.
type PlayerManaAbility struct {
	// LifeCost is the life each activation costs (CR 119.4: a player
	// can pay life only if their life total is at least the payment).
	LifeCost int `json:"lifeCost"`
	// Produced is the mana each activation adds, in the produced-mana
	// notation a card's mana ability uses ("{C}"). Fixed colours only:
	// a pick ("{W|U}") is refused at the grant, because nothing asks
	// one for a player-level ability.
	Produced string `json:"produced"`
}

// PlayerManaAbilityOffer is one live player mana ability as a reader
// sees it: what it costs, what it makes, where it came from and
// whether the player can pay for it now.
type PlayerManaAbilityOffer struct {
	// Source is the object that granted it (the Channel spell) and the
	// handle an activation names. Two grants from one source are one
	// offer; two Channels are two.
	Source uuid.UUID
	// Label is the attribution ("Channel").
	Label    string
	LifeCost int
	Produced string
	// Payable is false while the player cannot pay the life (CR 119.4,
	// or a life total that can't change, CR 119.8).
	Payable bool
}

// playerManaAbilityColors parses a fixed-colour produced string into
// one colour per mana. Nil when it is malformed or names a pick.
func playerManaAbilityColors(produced string) []string {
	slots, err := ParseProducedMana(produced)
	if err != nil || len(slots) == 0 {
		return nil
	}
	out := make([]string, 0, len(slots))
	for _, s := range slots {
		if len(s.Options) != 1 || s.DifferentColors() {
			return nil
		}
		out = append(out, s.Options[0])
	}
	return out
}

// GrantPlayerManaAbilityForEffect gives `player` the mana ability `ab`
// for duration d (until end of turn when d is zero). `source` is the
// granting object and `label` its attribution. A malformed ability
// grants nothing. Caller must hold g.mu (write), which a resolution
// frame already does.
func (g *Game) GrantPlayerManaAbilityForEffect(player uuid.UUID, ab PlayerManaAbility, label string, source uuid.UUID, d Duration) {
	if ab.LifeCost < 0 || playerManaAbilityColors(ab.Produced) == nil {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	grant := ab
	p.Statics = append(p.Statics, PlayerStatic{
		ManaAbility: &grant,
		Source:      source,
		Label:       label,
		Duration:    d,
	})
}

// PlayerManaAbilitiesLocked lists the live player mana abilities of
// `playerID`, one offer per granting source, in grant order. The view
// and the legal-move enumerator read it. Caller must hold g.mu.
func (g *Game) PlayerManaAbilitiesLocked(playerID uuid.UUID) []PlayerManaAbilityOffer {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	var out []PlayerManaAbilityOffer
	for _, s := range p.Statics {
		if s.ManaAbility == nil || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		dup := false
		for _, o := range out {
			if o.Source == s.Source {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		out = append(out, PlayerManaAbilityOffer{
			Source:   s.Source,
			Label:    s.Label,
			LifeCost: s.ManaAbility.LifeCost,
			Produced: s.ManaAbility.Produced,
			Payable:  g.CanPayLifeLocked(p, s.ManaAbility.LifeCost) && !g.LandManaOnlyLocked(playerID),
		})
	}
	return out
}

// playerManaAbilityLocked finds the live grant `source` made to p.
func (g *Game) playerManaAbilityLocked(p *Player, source uuid.UUID) (PlayerStatic, bool) {
	for _, s := range p.Statics {
		if s.ManaAbility != nil && s.Source == source && !g.durationExpiredLocked(s.Duration, false) {
			return s, true
		}
	}
	return PlayerStatic{}, false
}

// ActivatePlayerManaAbility activates the player mana ability that
// `source` granted `playerID`: pay its life, add its mana. Like any
// mana ability it uses no stack and resolves at once (CR 605.3b), so
// it is accepted whenever the actions layer accepts a mana activation.
//
// Refused, with nothing paid, when no live grant from `source` exists
// (it expired at cleanup, or never was), when the player cannot pay the
// life (CR 119.4: you can't pay more life than you have; CR 119.8), and
// under Word of Command's "only mana abilities of lands" (ADR 0166).
func (g *Game) ActivatePlayerManaAbility(playerID, source uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.activatePlayerManaAbilityLocked(playerID, source); err != nil {
		return err
	}
	// #2275 / CR 117.3c, as for ActivateManaAbility: the priority
	// holder who makes mana has taken an action.
	g.noteActionTakenLocked(playerID)
	return nil
}

func (g *Game) activatePlayerManaAbilityLocked(playerID, source uuid.UUID) error {
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	grant, ok := g.playerManaAbilityLocked(p, source)
	if !ok {
		return fmt.Errorf("%w: %s has no mana ability from that source this turn", ErrInvalidParam, p.Name)
	}
	if g.LandManaOnlyLocked(playerID) {
		return ErrCantActivate
	}
	ab := *grant.ManaAbility
	colors := playerManaAbilityColors(ab.Produced)
	if colors == nil {
		return ErrInvalidParam
	}
	if !g.CanPayLifeLocked(p, ab.LifeCost) {
		// CR 119.4: you can't pay more life than you have.
		return fmt.Errorf("%w: %s cannot pay %d life (CR 119.4)", ErrInvalidParam, p.Name, ab.LifeCost)
	}
	if ab.LifeCost > 0 {
		if err := g.PayLifeForEffect(source, playerID, ab.LifeCost); err != nil {
			return err
		}
		// Paying down to 0 is legal; the state-based actions end the
		// player after the mana is in the pool.
		defer g.runStateChecksLocked()
	}
	g.produceManaLocked(p, source, colors, nil, nil, 0, false, nil)
	return nil
}
