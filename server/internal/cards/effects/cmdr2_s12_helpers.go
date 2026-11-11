package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s12_helpers.go — shared card-side helpers for batch 12 (#5368).

// cmdr2S12ElementalSpectacle is Elemental Spectacle's resolution: one 5/5
// red and green Elemental per colour among the controller's permanents,
// then life equal to the creatures the controller now controls.
func cmdr2S12ElementalSpectacle(item *game.StackItem, ctx *Context) error {
	n := len(colorsAmongPermanentsYouControl(ctx.Game, item.Controller))
	if n > 0 {
		if err := (CreateToken{Template: TokenCard("5/5 red and green Elemental"), N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Amount: creaturesControlledBy(ctx.Game, item.Controller)}.Apply(ctx)
}

// cmdr2S12CreatureCountElementalToken is the green and white Elemental of
// Vernal Sovereign (Lorwyn Eclipsed Commander): "This token's power and
// toughness are each equal to the number of creatures you control." It is a
// catalog template, not a tokens_table.go row, because the
// characteristic-defining ability is a static the token always prints
// (CR 604.3, layer 7a). The token counts itself. Its slug is an on-disk
// identity: never renamed, never reused.
func cmdr2S12CreatureCountElementalToken() game.Card {
	return tokenFromCatalog(printedCreatureCountElementalToken)
}

func printedCreatureCountElementalToken() tokenTemplate {
	const text = "This token's power and toughness are each equal to the number of creatures you control."
	return tokenTemplate{
		Slug: "elemental-creature-count",
		Card: game.Card{
			Name:     "Elemental",
			TypeLine: "Token Creature — Elemental",
			Colors:   []string{"G", "W"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     text,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsCreature() {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: text,
	}
}

// cmdr2S12SoulstokeSacrifice is Incandescent Soulstoke's delayed trigger:
// sacrifice the creature it put onto the battlefield, only if it is still
// that object and still under the controller's control (CR 603.7c). An
// on-disk key.
var cmdr2S12SoulstokeSacrifice = game.DelayedBody("incandescent-soulstoke/sacrifice", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || !onBattlefield(g, p.Object.ID) || c.ObjectEpoch != p.Object.Epoch || c.Controller != item.Controller {
		return nil
	}
	return SacrificePermanent{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// cmdr2S12SoulstokeEntered is "That creature gains haste until end of turn.
// Sacrifice it at the beginning of the next end step", run once the card has
// landed. A declined or empty put makes no delayed trigger. An on-disk key.
var cmdr2S12SoulstokeEntered = PutFromHandTail("incandescent-soulstoke/entered", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Incandescent Soulstoke — haste"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label:  "Incandescent Soulstoke — sacrifice that creature",
		Body:   cmdr2S12SoulstokeSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

// cmdr2S12HordePlay is Horde of Notions' resolution: the targeted Elemental
// card, if it is still in the controller's graveyard, may be played from there
// without paying its mana cost, during this resolution (CR 608.2g).
func cmdr2S12HordePlay(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		z := g.FindCardZoneForEffect(t.ID)
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      ctx.Controller(),
			Zone:        game.ZoneGraveyard,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			AltCostKey:  game.AltCostKeyGrantedCost,
			Timing:      game.TimingFlash,
			LapseOnPass: game.LapseStaysInHand,
			Source:      ctx.Source(),
			SourceName:  "Horde of Notions",
			Label:       "Horde of Notions — play " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Horde of Notions", game.MayCastKeywordFreeCast, declineLeaveInGraveyard, c.InstanceID)
		return nil
	}
	return nil
}

// cmdr2S12EndRazePump is End-Raze Forerunners' enters trigger: other
// creatures its controller controls get +2/+2 and gain vigilance and trample
// until end of turn.
func cmdr2S12EndRazePump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	others := And(Creature(), YouControl(), OtherThan(item.SourceCardID))
	if err := (BoostUntilEOT{Match: others, Power: 2, Toughness: 2, Label: "End-Raze Forerunners — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Match: others, Keywords: []string{"vigilance", "trample"}, Label: "End-Raze Forerunners — vigilance and trample"}.Apply(ctx)
}

// cmdr2S12ElementalManaRestrictions is "Spend this mana only to cast
// Elemental spells or activate abilities of Elementals" (Smokebraider,
// Primal Beyond): one ManaRestrictAnyOf tag, a cast of an Elemental spell or
// an activation whose source is an Elemental (CR 106.6). A changeling counts
// (CR 702.73a). Eclipsed Realms' clause with the type fixed.
func cmdr2S12ElementalManaRestrictions() []string {
	tribe := game.ManaRestrictSubtype("Elemental")
	return []string{game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, tribe},
		[]string{game.ManaRestrictActivate, tribe},
	)}
}

// cmdr2S12IsElementalCard and cmdr2S12IsFaerieCard are the reveal tests of
// Primal Beyond, Flamekin Village and Secluded Glen: a card in hand with the
// subtype on its printed type line (no layer applies in a hand).
func cmdr2S12IsElementalCard(c game.Card) bool { return c.HasSubtypeWord("Elemental") }
func cmdr2S12IsFaerieCard(c game.Card) bool    { return c.HasSubtypeWord("Faerie") }
