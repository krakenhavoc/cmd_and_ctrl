package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivi Ornitier — Legendary Creature — Wizard {1}{U}{R}, 0/3:
//
//	"{0}: Add X mana in any combination of {U} and/or {R}, where X is
//	 Vivi Ornitier's power. Activate only during your turn and only
//	 once each turn.
//	 Whenever you cast a noncreature spell, put a +1/+1 counter on
//	 Vivi Ornitier and it deals 1 damage to each opponent."
//
// A 0/3 that does nothing the turn it lands and wins the game two
// turns later. The two halves are one engine: every noncreature
// spell grows Vivi and pings the table, and once a turn Vivi cashes
// that accumulated power for free mana to cast more of them. The
// first activation each turn is the ramp; the spells cast off it are
// what make the NEXT one bigger.
//
// The mana ability, component by component, because each one is a
// printed clause rather than a convenience:
//
//   - "{0}" is a cost with no components at all — ManaAbilityCost{}.
//     Vivi does not tap for it, so a summoning-sick Vivi could fire
//     it (and the once-per-turn gate is what stops that mattering),
//     and the auto-tapper never plans it, because an ability with no
//     {T} is not a tap source.
//   - "X … where X is Vivi Ornitier's power" is the SCALED shape of
//     ProducedFunc: the pipe slot repeated once per point of power.
//     Power is read at activation, post-layers and with counters
//     folded in (CurrentPower), so the +1/+1 counters the trigger
//     banks, an anthem and an Equipment all count. A Vivi at power 0
//     — which is every Vivi the turn it enters — produces nothing and
//     the activation is a no-op, exactly as X = 0 is printed.
//   - "in any combination of {U} and/or {R}" is why each slot is its
//     own pipe rather than one wide token: the controller answers a
//     colour pick per slot, so three power can be {U}{U}{R} and not
//     only three of one colour. NarrowToCommanderIdentity stays off —
//     the printed text names two colours itself and says nothing
//     about a commander's identity, so the offer is exactly {U} or
//     {R} whatever the commander is.
//   - "Activate only during your turn and only once each turn" is
//     TWO conditions, ANDed (AllConditions). Both are CR 602.1b
//     gates checked before anything is paid. The second is
//     ManaAbilityNotUsedThisTurn, which counts this permanent's
//     EventManaAbilityActivated in the turn's own slice of the event
//     log; Vivi has exactly one mana ability, so the helper's
//     per-permanent key is exact here.
//
// The trigger is "whenever YOU cast a NONCREATURE spell" — any
// noncreature spell, including one that is countered afterwards,
// since the trigger fires on the cast (CR 601.2i) and not on
// resolution. The counter goes on first and the damage follows, both
// from the resolving ability rather than from the spell. A Vivi that
// has left the battlefield by the time the ability resolves still
// deals the damage (the source is last-known information, CR
// 608.2h); only the counter is skipped, because there is nothing on
// the battlefield to put it on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0452be2a-e97a-4269-a9e3-b616265ecb2e",
		Name:         "Vivi Ornitier",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{},
			ProducedFunc: viviManaFromPower,
			Label:        "{0}: Add X mana in any combination of {U} and/or {R}, where X is Vivi Ornitier's power. Activate only during your turn and only once each turn",
			Condition:    AllConditions(DuringYourTurn(), ManaAbilityNotUsedThisTurn()),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Vivi Ornitier — +1/+1 counter and 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					if b15OnBattlefield(g, item.SourceCardID) {
						add := AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}
						if err := add.Apply(NewContext(g, item)); err != nil {
							return err
						}
					}
					return damageToEachOpponent(g, item, 1)
				}),
		},
	})
}

// viviManaFromPower is "X mana in any combination of {U} and/or {R},
// where X is Vivi Ornitier's power": one {U|R} slot per point of
// power, so the controller picks a colour for each of them
// independently.
//
// Read-only under g.mu, like every ProducedFunc. CurrentPower clamps
// at zero, so a Vivi shrunk below 0 produces nothing rather than
// asking strings.Repeat for a negative count.
func viviManaFromPower(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	return strings.Repeat("{U|R}", c.CurrentPower())
}
