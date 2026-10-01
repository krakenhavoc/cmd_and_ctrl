package game

import "github.com/google/uuid"

// cant_gain_life.go — "players can't gain life", read at one gate (ADR
// 0107 §5 decision 6, owner decision 5, #1880).
//
// THE RULE. CR 119.7: "If an effect says that a player can't gain life,
// that player can't make an exchange such that the player's life total
// would become higher; in that case, the exchange won't happen. … In
// addition, a cost that involves having that player gain life can't be
// paid, and a replacement effect that would replace a life gain event
// affecting that player won't do anything." CR 119.10: "If a player
// gains 0 life, no life gain event would occur".
//
// A RULES EFFECT, NOT A REPLACEMENT (CR 613.11), the sibling of "damage
// can't be prevented" (unpreventable_damage.go). It is asked at every
// life gain BEFORE the CR 614 window opens
// (changeLifeThroughReplacementsLocked), so an "if you would gain life"
// replacement — Rhox Faithmender's doubler, Sulfuric Vortex's "gains no
// life instead" — never sees a gain that can't happen, and nothing is
// gained: no EventChangeLife, so "whenever you gain life" does not
// trigger. Every writer of a life total already runs that one window
// (#482): a catalog GainLife, CR 702.15b lifelink, a drain's gain half,
// a "life total becomes N" that would raise it, the sandbox verb. None
// of them learns this rule. Lifelink damage is still DEALT; only the
// gain is stopped.
//
// "Your life total can't change" (life_lock.go, CR 119.8 too) is a
// separate, wider prohibition and keeps its own built-in.
//
// THE THREE SOURCES, the first yes winning:
//
//  1. Battlefield statics — "Players can't gain life" (Leyline of
//     Punishment, Sunspine Lynx, Giant Cindermaw), "Your opponents
//     can't gain life" (Erebos, God of the Dead, Archfiend of Despair)
//     and "Enchanted player can't gain life" (Grievous Wound). Read
//     live, keyed by CatalogAbilityKey (CR 613.1f).
//  2. Turn grants — Skullcrack's "players can't gain life this turn",
//     Atarka's Command's "your opponents … this turn": a ModCantGainLife
//     ScopedEffect, swept at cleanup (CR 514.2).
//  3. The rest of the game — Screaming Nemesis's and Stigma Lasher's
//     "that player can't gain life for the rest of the game": the same
//     kind with Player set and an indefinite duration (CR 611.2a).
//     It outlives the source, which is why it has to be stored.

// CantGainLifeWhose is which players a printed "can't gain life" static
// covers.
type CantGainLifeWhose uint8

const (
	// CantGainLifeEveryone is "Players can't gain life".
	CantGainLifeEveryone CantGainLifeWhose = iota
	// CantGainLifeOpponents is "Your opponents can't gain life": every
	// opponent of the static's controller.
	CantGainLifeOpponents
	// CantGainLifeEnchantedPlayer is "Enchanted player can't gain life"
	// (Grievous Wound): the player the Aura is attached to.
	CantGainLifeEnchantedPlayer
)

// CantGainLifeStatic is one printed "can't gain life" static on a
// permanent. Catalog data, never stored.
type CantGainLifeStatic struct {
	// Label is the clause as printed.
	Label string
	// Whose is which players it covers.
	Whose CantGainLifeWhose
}

// CatalogCantGainLife returns the "can't gain life" statics a permanent
// with the given catalog key has. carddef.go sets it from
// CardDef.CantGainLife; a game-package test may stub it.
var CatalogCantGainLife func(key string) []CantGainLifeStatic

// playerCantGainLifeLocked is THE reader: may this player gain life right
// now? A nil or eliminated player is not covered (the life tail reports
// its absence on its own).
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) playerCantGainLifeLocked(p *Player) bool {
	if p == nil {
		return false
	}
	// 1. The battlefield statics.
	if CatalogCantGainLife != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			key := CatalogAbilityKey(*src)
			if key == "" {
				continue
			}
			for _, s := range CatalogCantGainLife(key) {
				if cantGainLifeStaticCovers(s.Whose, src, p.ID) {
					return true
				}
			}
		}
	}
	// 2 and 3. The stored grants.
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantGainLife {
				continue
			}
			switch e.Scope {
			case ScopeGame:
				if m.Player == uuid.Nil || m.Player == p.ID {
					return true
				}
			default:
				if scopeCoversPlayer(e.Scope, e.Controller, p.ID) {
					return true
				}
			}
		}
	}
	return false
}

// cantGainLifeStaticCovers reports whether a static on `src` covers the
// player `player`.
func cantGainLifeStaticCovers(whose CantGainLifeWhose, src *Card, player uuid.UUID) bool {
	switch whose {
	case CantGainLifeEveryone:
		return true
	case CantGainLifeOpponents:
		return src.Controller != uuid.Nil && player != src.Controller
	case CantGainLifeEnchantedPlayer:
		return src.AttachedTo.Kind == TargetPlayer && src.AttachedTo.ID == player
	}
	return false
}

// PlayerCantGainLifeLocked is the exported surface over the reader, for
// the protocol projection's badge and for card files that have to ask
// before an exchange of life totals (CR 119.7).
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerCantGainLifeLocked(p *Player) bool {
	return g.playerCantGainLifeLocked(p)
}

// lifeGainForbiddenLocked is the gate changeLifeThroughReplacementsLocked
// asks before the CR 614 window: is this a life GAIN for a player who
// can't gain life? A loss, and a change of zero (CR 119.10), are never
// stopped here.
//
// Caller must hold g.mu.
func (g *Game) lifeGainForbiddenLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventLife || ev.LifeDelta <= 0 {
		return false
	}
	return g.playerCantGainLifeLocked(g.playerByIDLocked(ev.LifePlayer))
}

// PlayersCantGainLifeForEffect registers "<players> can't gain life"
// from a resolving spell or ability, for duration `d`:
//
//   - CantGainLifeEveryone: every player (Skullcrack, Call In a
//     Professional).
//   - CantGainLifeOpponents: the opponents of `controller` (Atarka's
//     Command, Roiling Vortex), read live as a rule (CR 611.2c).
//   - CantGainLifeEnchantedPlayer is not a duration effect and is
//     refused; for one player use PlayerCantGainLifeForEffect.
//
// Reports whether a record was written. Caller must hold g.mu (write).
func (g *Game) PlayersCantGainLifeForEffect(sourceID, controller uuid.UUID, whose CantGainLifeWhose, d Duration, label string) bool {
	switch whose {
	case CantGainLifeEveryone:
		return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
			[]Mod{{Kind: ModCantGainLife}}, d, label)
	case CantGainLifeOpponents:
		if controller == uuid.Nil {
			return false
		}
		return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeOpponentsAndTheirCreatures, controller,
			[]Mod{{Kind: ModCantGainLife}}, d, label)
	}
	return false
}

// PlayerCantGainLifeForEffect registers "<player> can't gain life" for
// duration `d` — Screaming Nemesis's and Stigma Lasher's "for the rest of
// the game" is IndefiniteDuration() (CR 611.2a). Registers nothing for a
// player who is not seated. Caller must hold g.mu (write).
func (g *Game) PlayerCantGainLifeForEffect(sourceID, player uuid.UUID, d Duration, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantGainLife, Player: player}}, d, label)
}

// GainNoLifeThisTurnForEffect is "if <player> would gain life this turn,
// that player gains no life instead" (Flames of the Blood Hand): a CR 614
// replacement on the life window (ModGainNoLife), not a "can't". Caller
// must hold g.mu (write).
func (g *Game) GainNoLifeThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModGainNoLife, Player: player}}, g.UntilEndOfTurnDuration(), label)
}
