package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b33_helpers.go — shared pieces for Star Trek Commander batch 33
// (#4974). Names are prefixed trekB33 so they cannot collide with another
// batch's.

// --- Tribble token -----------------------------------------------------------

// trekB33TribbleSlug is the catalog identity of the Tribble token.
const trekB33TribbleSlug = "tribble"

// printedTrekB33TribbleToken is the 1/1 white Tribble creature token with
// "This token can't block." (Tribble Cluster). A catalog template and not
// a tokens_table row because the can't-block clause is a static the token
// always prints (the Phyrexian Mite's shape). The token card in the dump
// (ttrc 2) lists no colour; the oracle text of the card that makes it says
// white, and the template follows the oracle text (CR 111.4), as The
// Eagles Are Coming!'s Bird Soldier does.
func printedTrekB33TribbleToken() tokenTemplate {
	return tokenTemplate{
		Slug: trekB33TribbleSlug,
		Card: game.Card{
			Name:      "Tribble",
			TypeLine:  "Token Creature — Tribble",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// TribbleToken is the Tribble's card value for CreateToken.
func TribbleToken() game.Card { return tokenFromCatalog(printedTrekB33TribbleToken) }

// --- "create X <tokens>, where X is its power" --------------------------------

// trekB33TokensEqualToItsPower is a dies trigger's "create X <token>,
// where X is its power" (Stratagem Emulator, Tribble Cluster). The power
// is the creature's as it last existed on the battlefield (CR 603.10a,
// CR 608.2h), so counters and pumps it died with count. Zero or less
// creates nothing.
func trekB33TokensEqualToItsPower(token func() game.Card) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		info, ok := ctx.SourcePermanent()
		if !ok || info.Power <= 0 {
			return nil
		}
		return CreateToken{Template: token(), N: info.Power}.Apply(ctx)
	}
}

// --- Spot, Felis catus ----------------------------------------------------------

// trekB33SpotOrAnotherNontokenLegendaryEntered is "Spot or another nontoken
// legendary permanent you control enters": Spot himself, token or not, or
// another permanent under his controller's control that is legendary and
// not a token.
func trekB33SpotOrAnotherNontokenLegendaryEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && !IsToken(c) && c.IsLegendary()
}

// --- Merciful Mandate -------------------------------------------------------------

// trekB33EndTheCombatPhase is the answer to Merciful Mandate's "you may end
// the combat phase": a yes ends it (CR 724.2). The spell is still on the
// stack while the question is open (the resolution tail, #3613), so CR
// 724.2b exiles it with everything else. Outside a combat phase it does
// nothing (CR 724.2g). An on-disk key.
var trekB33EndTheCombatPhase = MayYes("confirm/merciful-mandate-end-combat", func(ctx *Context, _ game.ConfirmAnswer) error {
	ctx.Game.EndCombatPhaseForEffect()
	return nil
})

// --- Right of Vengeance -------------------------------------------------------------

var (
	// "Whenever a creature deals combat damage to you": one instance for
	// each creature, carrying it (with its epoch) as the payload.
	trekB33CreatureHitsYouCondition = game.DelayedEachCondition("damage/a-creature-combat-to-you", trekB33CreatureCombatDamageToYou)

	// "Destroy that creature and create a 2/1 red Klingon Warrior creature
	// token with haste."
	trekB33DestroyItAndMakeAKlingonBody = game.SimpleDelayedBody("destroy/that-creature-and-klingon-warrior", trekB33DestroyItAndMakeAKlingon)
)

// trekB33CreatureCombatDamageToYou is the creature that dealt combat damage
// to the delayed trigger's controller. Only creatures deal combat damage
// (CR 510.1a), and damage of 0 is no damage (CR 120.8).
func trekB33CreatureCombatDamageToYou(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target != dt.Controller {
		return nil
	}
	return []uuid.UUID{ev.Source}
}

// trekB33DestroyItAndMakeAKlingon destroys the payload creature if it is
// still that object on the battlefield (CR 603.7c), then creates the
// Klingon Warrior whether or not it was destroyed.
func trekB33DestroyItAndMakeAKlingon(g *game.Game, item *game.StackItem) error {
	if err := destroyThePayloadPermanent(g, item); err != nil {
		return err
	}
	return CreateToken{Template: TokenCard("2/1 red Klingon Warrior with haste"), N: 1}.Apply(NewContext(g, item))
}
