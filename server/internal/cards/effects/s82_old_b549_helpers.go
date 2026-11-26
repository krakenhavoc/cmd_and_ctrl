package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b549_helpers.go — Lorwyn, life gain and loss (S82 older sets,
// batch 549, #5593): the shared shapes of Judge of Currents and Bog-Strider
// Ash.

// S82OldB549MerfolkYouControlBecameTapped is "a Merfolk you control becomes
// tapped": the tap watch of Magda, Brazen Outlaw over Merfolk, read through
// the source's text (CR 612.2). An attacker that is now tapped became tapped;
// a vigilance attacker did not. Watch game.EventTapCard and game.EventAttack.
func S82OldB549MerfolkYouControlBecameTapped(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller || !c.IsCreature() || !c.HasSubtype(source.TextSubtype("Merfolk")) {
		return false
	}
	return ev.Kind != game.EventAttack || c.Tapped
}

// S82OldB549BogStriderAshPaid is "you gain 2 life", run once {G} is paid.
// An on-disk key: never renamed, never reused.
var S82OldB549BogStriderAshPaid = PayUnlessBranches("pay-unless/s82-old-b549-bog-strider-ash",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return GainLife{Amount: 2}.Apply(ctx)
	}, nil)

// S82OldB549BogStriderAshOffer is the trigger's resolution: its controller may
// pay {G}.
func S82OldB549BogStriderAshOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{G}",
		Question: "Bog-Strider Ash — pay {G} to gain 2 life?",
		Then:     S82OldB549BogStriderAshPaid,
	}.Apply(NewContext(g, item))
}
