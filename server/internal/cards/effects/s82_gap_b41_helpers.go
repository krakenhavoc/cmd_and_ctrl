package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB41FreerunningReturn is "Freerunning—Return a <creature> you control
// to its owner's hand" (Escape Detection): the freerunning alternative cost
// (CR 702.173a) paid with a bounce instead of mana. It keeps the
// "freerunning" key and the Assassin-or-commander gate of Freerunning, and
// adds the return component Daze's alternative cost uses.
func gapB41FreerunningReturn(label string, from *game.TargetSpec, payLabel string) game.AlternativeCost {
	alt := Freerunning("", "Assassin")
	alt.Label = label
	alt.ReturnToHand = from
	alt.PayLabel = payLabel
	return alt
}

var (
	// gapB41AssassinYouControlAttacks is "whenever an Assassin you control
	// attacks": an EventAttack declared by the delayed trigger's controller
	// whose attacker has the creature type in the params' Name (the card's
	// "Assassin" read through its text, CR 612.2) as it attacks (layered
	// subtypes).
	// Without OncePerBatch it fires once for each Assassin (CR 603.2c).
	gapB41AssassinYouControlAttacks = game.DelayedCondition("attack/an-assassin-you-control-attacks",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
			if ev.Kind != game.EventAttack || ev.Actor != dt.Controller || p.Name == "" {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.HasSubtype(p.Name)
		})

	// gapB41AssassinAttackingBody is Origin of the Hidden Ones' delayed
	// trigger: a 1/1 black Assassin with menace, tapped and attacking.
	gapB41AssassinAttackingBody = game.SimpleDelayedBody("origin-of-the-hidden-ones/assassin-tapped-and-attacking", gapB41AssassinAttacking)
)

// gapB41AssassinAttacking creates one tapped and attacking Assassin. Its
// controller chooses what it attacks (CR 508.4), the attacker's target
// offered first. It was never declared as an attacker, so it triggers
// nothing that watches attacks, this trigger included.
func gapB41AssassinAttacking(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	return CreateTokensAttackingYourChoice{
		Template: TokenCard("1/1 black Assassin with menace"),
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Origin of the Hidden Ones",
	}.Apply(NewContext(g, item))
}
