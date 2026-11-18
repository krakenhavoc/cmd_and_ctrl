package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set SNC, batch 17 (#5491): attack triggers and combat.

// snc17TwoMulticoloredPermanents is "you control two or more multicolored
// permanents" (Scuttling Butler), read on the layered colours.
var snc17TwoMulticoloredPermanents = CondBoard{
	Name: "two or more multicolored permanents you control",
	Test: func(ctx *Context) bool {
		return len(MatchingBattlefield(ctx, And(YouControl(), Multicolored()))) >= 2
	},
}

// snc17NoDevilToken is "you don't control a Devil token" (Maestros
// Diabolist), read on the layered subtypes.
var snc17NoDevilToken = CondBoard{
	Name: "you don't control a Devil token",
	Test: func(ctx *Context) bool {
		return len(MatchingBattlefield(ctx, And(YouControl(), IsTokenPredicate(), OfSubtype("Devil")))) == 0
	},
}

// snc17Parley is Phabine's parley (cmdr2S49Parley's shape, which counts only
// the nonland cards): each player in turn order reveals the top card of
// their library, `reward` is handed how many were lands and how many were
// not, then each player draws a card. A player with an empty library
// reveals nothing.
func snc17Parley(g *game.Game, source uuid.UUID, label string, reward func(lands, nonland int) error) error {
	players := apnapPlayers(g)
	lands, nonland := 0, 0
	for _, p := range players {
		for _, id := range g.RevealTopOfLibraryForEffect(p, source, 1, label) {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			if c.IsLand() {
				lands++
			} else {
				nonland++
			}
		}
	}
	if err := reward(lands, nonland); err != nil {
		return err
	}
	for _, p := range players {
		if err := g.DrawNForEffect(p, 1); err != nil {
			return err
		}
	}
	return nil
}

var (
	// sncIndulgeAttackCondition is "whenever a creature you control attacks":
	// an EventAttack declared by the delayed trigger's controller. Without
	// OncePerBatch it fires once for each creature (CR 603.2c).
	sncIndulgeAttackCondition = game.DelayedCondition("attack/a-creature-you-control-attacks",
		func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventAttack && ev.Actor == dt.Controller
		})

	// sncIndulgeCitizenBody is Indulge's delayed trigger: a 1/1 green and
	// white Citizen, tapped and attacking.
	sncIndulgeCitizenBody = game.SimpleDelayedBody("indulge-excess/citizen-tapped-and-attacking", sncIndulgeCitizen)
)

// sncIndulgeCitizen creates one tapped and attacking Citizen. Its controller
// chooses what it attacks (CR 508.4), the attacker's target offered first. It
// was never declared as an attacker, so it triggers nothing that watches
// attacks, this trigger included.
func sncIndulgeCitizen(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	return CreateTokensAttackingYourChoice{
		Template: TokenCard("1/1 green and white Citizen"),
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Indulge",
	}.Apply(NewContext(g, item))
}

// sncCreaturesYouControlledThatDealtCombatDamageToAPlayer is how many
// distinct creatures `controller` controlled as they dealt combat damage to
// a player this turn (Excess), read off this turn's damage events: a combat
// event's Actor is the damaging creature's controller as it dealt the damage
// (CR 608.2h), so a creature that has died, left or changed hands still
// counts for the player who controlled it then, and one creature that dealt
// damage to two players, or in two combat damage steps, counts once.
func sncCreaturesYouControlledThatDealtCombatDamageToAPlayer(g *game.Game, controller uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Actor != controller {
			continue
		}
		if g.PlayerByIDForEffect(ev.Target) == nil {
			continue
		}
		seen[ev.Source] = true
	}
	return len(seen)
}
