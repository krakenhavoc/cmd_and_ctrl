package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b154_helpers.go — GAP batch 154 (Jumpstart 2022 creatures,
// #5948): what two or more cards of the batch share, and the Banana token.

// GapB154BananaToken is the Banana Kibo, Uktabi Prince makes: a colorless
// artifact token named Banana with "{T}, Sacrifice this token: Add {R} or
// {G}. You gain 2 life." It has no printed token card: its characteristics
// follow the oracle text. Gaining the life is part of the mana ability
// (CR 605.1a: it has no target and adds mana), declared so the auto-tapper
// plans it. Its slug is an on-disk identity and is never renamed.
func GapB154BananaToken() game.Card { return TokenFromCatalog(printedGapB154BananaToken) }

func printedGapB154BananaToken() TokenTemplate {
	return TokenTemplate{
		Slug: "gap-b154-banana",
		Card: game.Card{
			Name:     "Banana",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:           true,
			SacrificeCost:     true,
			Produced:          "{R|G}",
			Rider:             GainLifeRider(2),
			RiderSelfLifeGain: 2,
			Label:             "{T}, Sacrifice: Add {R} or {G}. You gain 2 life",
		}},
		Text: "{T}, Sacrifice this token: Add {R} or {G}. You gain 2 life.",
	}
}

// GapB154YouControlADamageSourceDealtDamageToYou is "a source you control
// deals damage to you" (Auntie Blyte): one damage event to the source's
// controller, the player and not their permanents, from a source its
// controller controls (the source itself, a Blyte that burns its own
// controller, included). The damage amount is the event's.
func GapB154SourceYouControlDealtDamageToYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	return B17DamageSourceController(ev, g) == source.Controller
}

// GapB154SnowPermanentYouControlEnters is "another snow permanent you
// control enters" (Isu the Abominable): any permanent card or token with the
// snow supertype entering under the source's control, read as it entered.
func GapB154AnotherSnowPermanentEntersUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSupertype("Snow")
}

// GapB154GreatSnowPay is "you may pay {G}, {W}, or {U}": one mana of any of
// the three colors, paid as a single three-way hybrid cost.
const GapB154GreatSnowPay = "{G/W/U}"
