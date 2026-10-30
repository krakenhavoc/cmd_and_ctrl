package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc10_helpers.go — shared bodies for Marvel Super Heroes Commander,
// batch 10 (#4455).

// battalionAttack is battalion's trigger condition: the source's own
// attack declaration with at least two OTHER creatures its controller
// controls attacking in the same declaration (Boros Strike-Captain,
// General Thunderbolt Ross).
func battalionAttack(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclared(ev, source) && attackersYouControlOtherThan(g, source.Controller, source.InstanceID) >= 2
}

// landsYouControlPT is the characteristic-defining ability "<this>'s
// power and toughness are each equal to the number of lands you control"
// (CR 604.3, layer 7a), counted live as the layers are recomputed
// (Ashaya, Flora Colossus). The label is the clause as the card prints it.
func landsYouControlPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		Label:     label,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := b10LandsControlled(g, source.Controller)
			c.Power, c.Toughness = n, n
		},
	}
}

// msc10HexproofTheLegalTarget gives hexproof until end of turn to the
// ability's still-legal target (CR 608.2b: one that left in response is
// skipped) — Foggy Nelson's second sentence.
func msc10HexproofTheLegalTarget(ctx *Context) error {
	for _, id := range legalTargetCards(ctx.Item, ctx.Game) {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"hexproof"},
			Label:    "Foggy Nelson — hexproof until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// msc10ExileThenItsControllerGainsItsPower is Captain Marvel's first
// ability: exile the still-legal target creature, then its controller
// gains life equal to its power. Controller and power are read before the
// exile, the card's last state on the battlefield (CR 608.2h).
func msc10ExileThenItsControllerGainsItsPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCards(item, g) {
		card, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		controller, power := card.Controller, card.CurrentPower()
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if err := (GainLife{Player: controller, Amount: power}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// msc10CreatureExiledFromTheBattlefield is "whenever a creature other
// than <this> is exiled from the battlefield": a battlefield exit bound
// for exile by a creature that was a creature as it last existed there.
func msc10CreatureExiledFromTheBattlefield(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneExile || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftAsType(ev, c, "creature")
}
