package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// power_up.go — #4527, CR 702.193: the vocabulary a card file writes a
// power-up ability in. The engine half is game/power_up.go.
//
//	Activated: []ActivatedAbility{
//	    PowerUp("{4}{W}: Put two +1/+1 counters on this creature.",
//	        ManaCost("{4}{W}"), putPlusOneCountersOnSource(2)),
//	},
//
// "Power-up — [cost]: [effect]. (Activate each power-up ability only
// once. Reduce the cost by its mana cost if it entered this turn.)" is
// two rules, and a card file must not assemble them: the constructor
// sets the one bit, and the engine reads the activation record and
// prices the reduction in the places every activation already goes
// through (game.Game.PowerUpUsed, the CR 601.2f pass).

// powerUpPrefix is the printed ability word, with the dash the label
// and the oracle text both use.
const powerUpPrefix = "Power-up — "

// PowerUp builds a power-up ability. `text` is the printed line AFTER
// the keyword, "[cost]: [effect]." — the constructor prepends
// "Power-up — " so the label reads as printed, which is what the oracle
// fixture compares.
func PowerUp(text string, cost game.AbilityCost, effect Effect) ActivatedAbility {
	return ActivatedAbility{
		Label:   powerUpPrefix + text,
		Cost:    cost,
		PowerUp: true,
		Effect:  effect,
	}
}

// powerUpAnswering is PowerUp with its ADR 0142 declaration: a row that
// only grows its source (counters), so the one thing it can do in
// response is pump (game.AnswerPump), read by smart autopass and the bot.
// Every power-up row so far is one; a row that does something else
// builds PowerUp and sets its own Purpose.
func powerUpAnswering(text string, cost game.AbilityCost, effect Effect) ActivatedAbility {
	ab := PowerUp(text, cost, effect)
	ab.Purpose.Answers = game.AnswerPump
	return ab
}

// YourPowerUpAbilitiesActivateAnAdditionalTime — "Each power-up ability
// of permanents you control can be activated an additional time" (Wonder
// Man, Hollywood Hero). It ADDS to the printed once rather than replacing
// it, so two of them make it three: the engine sums every static that
// reaches the ability (game.Game.PowerUpLimitFor).
//
// It reaches the permanents its controller controls, whichever turn it
// is, and ends when its permanent leaves the battlefield (CR 113.6):
// nothing is stored, so an ability activated twice under it is simply
// spent afterwards.
func YourPowerUpAbilitiesActivateAnAdditionalTime(label string) game.PowerUpExtra {
	return game.PowerUpExtra{
		Label: label,
		Times: 1,
		Applies: func(_ *game.Game, source game.Card, grantor game.Card) bool {
			return source.Controller == grantor.Controller
		},
	}
}

// PowerUpAbilitiesCantBeActivatedDuringThatTurn is Kang the Conqueror's
// "During that turn, power-up abilities can't be activated": the rules
// effect an extra turn carries (TakeExtraTurn.DuringThatTurn), registered
// as the turn begins and gone at its cleanup step. Game-wide, so it
// bans every player's power-up abilities, Kang's controller's included.
func PowerUpAbilitiesCantBeActivatedDuringThatTurn() []game.Mod {
	return []game.Mod{{Kind: game.ModCantActivatePowerUp}}
}

// APowerUpAbility — the activation event is of a power-up ability
// ("whenever you activate a power-up ability", Marvel Boy, Noh-Varr).
// Reads the bit the announcement stamped (game.Event.PowerUp), for the
// reason ABoastAbility does.
func APowerUpAbility(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.PowerUp
}

// APowerUpAbilityCost — the ability being priced is a power-up ability
// ("Power-up abilities of other creatures you control cost {3} less to
// activate", Hulk, Gamma Goliath).
func APowerUpAbilityCost() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && q.Ability.PowerUp
	}
}

// OfAnotherCreatureYouControl — the ability's source is a creature
// other than the modifier's source, controlled by the modifier's
// controller ("abilities of other creatures you control").
func OfAnotherCreatureYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.InstanceID != q.Source.InstanceID &&
			q.Card.Controller == q.Source.Controller &&
			q.Card.IsCreature()
	}
}

// TheFirstPowerUpDuringEachOfYourTurns — the activator is the modifier's
// controller, it is their turn, and they have activated no power-up
// ability yet this turn (game.Game.PowerUpActivationsThisTurn, written
// at the announce, so a countered power-up still counts as the first).
func TheFirstPowerUpDuringEachOfYourTurns() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Controller == q.Source.Controller &&
			IsYourTurn(q.Game, q.Controller) &&
			q.Game.PowerUpActivationsThisTurn(q.Controller) == 0
	}
}

// PayZeroRatherThanThePowerUpCost is Advancing the Spirit's "You may pay
// {0} rather than pay the power-up cost of the first power-up ability
// you activate during each of your turns": a CR 118.9 alternative cost
// on an activation (game.CostZeroAlternative), which the engine takes
// for the activator because declining can never help them.
func PayZeroRatherThanThePowerUpCost(label string) game.CostModifier {
	return game.CostModifier{
		Kind:        game.CostZeroAlternative,
		Label:       label,
		Activations: true,
		AppliesTo:   allOf([]CostPredicate{APowerUpAbilityCost(), TheFirstPowerUpDuringEachOfYourTurns()}),
	}
}
