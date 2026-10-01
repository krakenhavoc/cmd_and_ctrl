package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// unpreventable_damage.go — the catalog's vocabulary for ADR 0107 §5
// (#1853, #1880): "damage can't be prevented" (CR 615.12) and "players
// can't gain life" (CR 119.7). The engine half, and the reason each
// form lives where it does, is game/unpreventable_damage.go and
// game/cant_gain_life.go.
//
//	DealDamage{…, CantBePrevented: true}               // Combust: "The damage can't be prevented."
//	DamageCantBePreventedThisTurn{}                     // Skullcrack: "Damage can't be prevented this turn."
//	Spec.DamageCantBePrevented: DamageCantBePreventedStatic()   // Leyline of Punishment
//	Spec.CantGainLife: PlayersCantGainLife()             // Leyline of Punishment
//	PlayersCantGainLifeThisTurn{}                        // Skullcrack: "Players can't gain life this turn."
//	PlayerCantGainLifeForRestOfGame{Player: p}           // Screaming Nemesis

// --- statics ---------------------------------------------------------

// DamageCantBePreventedStatic is "Damage can't be prevented." (Leyline
// of Punishment, Sunspine Lynx).
func DamageCantBePreventedStatic() []game.UnpreventableDamageStatic {
	return []game.UnpreventableDamageStatic{{Label: "Damage can't be prevented.", Scope: game.UnpreventableAll}}
}

// CombatDamageCantBePrevented is "Combat damage can't be prevented."
// (Frenzied Baloth).
func CombatDamageCantBePrevented() []game.UnpreventableDamageStatic {
	return []game.UnpreventableDamageStatic{{Label: "Combat damage can't be prevented.", Scope: game.UnpreventableCombat}}
}

// CombatDamageByYourCreaturesCantBePrevented is "Combat damage that
// would be dealt by creatures you control can't be prevented." (Questing
// Beast).
func CombatDamageByYourCreaturesCantBePrevented() []game.UnpreventableDamageStatic {
	return []game.UnpreventableDamageStatic{{
		Label: "Combat damage that would be dealt by creatures you control can't be prevented.",
		Scope: game.UnpreventableCombatByYourCreatures,
	}}
}

// DamageByThisCantBePrevented is "Damage that would be dealt by this
// creature can't be prevented." (Excruciator, Malignus).
func DamageByThisCantBePrevented() []game.UnpreventableDamageStatic {
	return []game.UnpreventableDamageStatic{{
		Label: "Damage that would be dealt by this creature can't be prevented.",
		Scope: game.UnpreventableByThis,
	}}
}

// PlayersCantGainLife is "Players can't gain life." (Leyline of
// Punishment, Giant Cindermaw, Havoc Festival).
func PlayersCantGainLife() []game.CantGainLifeStatic {
	return []game.CantGainLifeStatic{{Label: "Players can't gain life.", Whose: game.CantGainLifeEveryone}}
}

// OpponentsCantGainLife is "Your opponents can't gain life." (Erebos,
// God of the Dead, Archfiend of Despair, Knight of Dusk's Shadow).
func OpponentsCantGainLife() []game.CantGainLifeStatic {
	return []game.CantGainLifeStatic{{Label: "Your opponents can't gain life.", Whose: game.CantGainLifeOpponents}}
}

// EnchantedPlayerCantGainLife is "Enchanted player can't gain life."
// (Grievous Wound).
func EnchantedPlayerCantGainLife() []game.CantGainLifeStatic {
	return []game.CantGainLifeStatic{{Label: "Enchanted player can't gain life.", Whose: game.CantGainLifeEnchantedPlayer}}
}

// --- resolution actions ----------------------------------------------

// DamageCantBePreventedThisTurn is "Damage can't be prevented this
// turn." (Skullcrack, Stomp, Flaring Pain): a rule grant until cleanup
// (CR 514.2) that covers every damage event, including damage from
// sources that did not exist when it resolved (CR 611.2c).
//
// Put it BEFORE the card's own damage in a Sequence: the printed order
// is the order the effects begin, and the damage that follows is
// covered.
type DamageCantBePreventedThisTurn struct {
	// Label is attribution for the log; defaults to the clause.
	Label string
}

func (d DamageCantBePreventedThisTurn) Apply(ctx *Context) error {
	label := d.Label
	if label == "" {
		label = "Damage can't be prevented this turn"
	}
	ctx.Game.DamageCantBePreventedThisTurnForEffect(ctx.Source(), label)
	return nil
}

// PlayersCantGainLifeThisTurn is "Players can't gain life this turn."
// (Skullcrack, Call In a Professional) — or, with Opponents, "Your
// opponents can't gain life this turn." (Atarka's Command, Roiling
// Vortex), "your" being the resolving object's controller.
type PlayersCantGainLifeThisTurn struct {
	Opponents bool
	Label     string
}

func (p PlayersCantGainLifeThisTurn) Apply(ctx *Context) error {
	whose, label := game.CantGainLifeEveryone, "Players can't gain life this turn"
	if p.Opponents {
		whose, label = game.CantGainLifeOpponents, "Your opponents can't gain life this turn"
	}
	if p.Label != "" {
		label = p.Label
	}
	ctx.Game.PlayersCantGainLifeForEffect(ctx.Source(), ctx.Controller(), whose, ctx.Game.UntilEndOfTurnDuration(), label)
	return nil
}

// PlayerCantGainLifeForRestOfGame is "<that player> can't gain life for
// the rest of the game" (Screaming Nemesis, Stigma Lasher): stored on
// the game with no end (CR 611.2a), so it outlives its source.
type PlayerCantGainLifeForRestOfGame struct {
	Player uuid.UUID
	Label  string
}

func (p PlayerCantGainLifeForRestOfGame) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = "Can't gain life for the rest of the game"
	}
	ctx.Game.PlayerCantGainLifeForEffect(ctx.Source(), p.Player, game.IndefiniteDuration(), label)
	return nil
}
