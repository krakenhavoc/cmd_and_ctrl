package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b45_helpers.go — shared bodies for GAP batch 45 (#5948).

// gapB45GrantsDemonstrate builds the trigger for a permanent that gives
// demonstrate to spells its controller casts (Silverquill Lecturer:
// "Creature spells you cast have demonstrate"), the demonstrate sibling of
// socB32GrantsStorm. A spell that has demonstrate gets the keyword's
// trigger when it is cast (CR 702.144a), so the grant is a battlefield
// trigger watching the cast; the item carries the spell as Params.Object
// and the "you may" is asked as it resolves.
func gapB45GrantsDemonstrate(label string, when func(spell game.Card, source *game.Card, g *game.Game) bool) game.TriggeredAbility {
	t := game.TriggeredAbility{
		Keyword:   KeywordDemonstrate,
		Key:       label,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: socB32YouCastASpellThat(when),
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label+" — demonstrate")
			item.Controller, item.Owner = ev.Actor, ev.Actor
			item.Params.Object = game.ObjectRef{ID: ev.CardID}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return gapB45DemonstrateSpell(g, item, item.Params.Object.ID)
		},
	}
	return Optional(t, "Demonstrate — copy this spell? If you do, choose an opponent to also copy it.")
}

// gapB45DemonstrateSpell is demonstrateEffect for a spell the item names
// rather than is.
func gapB45DemonstrateSpell(g *game.Game, item *game.StackItem, spell uuid.UUID) error {
	ctx := NewContext(g, item)
	if err := (CopySpell{StackID: spell, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Demonstrate — choose an opponent to also copy the spell",
		ThenKey:  demonstrateOpponentCopies,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}

// gapB45ExcessDealtTo is the total excess damage (CR 120.10) dealt to the
// permanent `id` so far this turn, summed over the turn's damage events.
// Caller holds the game lock (a trigger's condition or effect).
func gapB45ExcessDealtTo(g *game.Game, id uuid.UUID) int {
	total := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventDealDamage && ev.Target == id {
			total += ev.ExcessDamage
		}
	}
	return total
}

// gapB45TarmogoyfToken is the Tarmogoyf creature token Tarmogoyf Nest
// makes: a green Lhurgoyf */1+* with the card's characteristic-defining
// ability. The token has no oracle ID, so the ability is registered under
// the token's key (token_catalog.go); it is appended to the catalog's list
// from this file's init, which runs before token_catalog.go's own.
func gapB45TarmogoyfToken() game.Card { return tokenFromCatalog(printedGapB45TarmogoyfToken) }

// printedGapB45TarmogoyfToken is the token as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func printedGapB45TarmogoyfToken() tokenTemplate {
	return tokenTemplate{
		Slug: "tarmogoyf-lhurgoyf",
		Card: game.Card{
			Name:      "Tarmogoyf",
			TypeLine:  "Token Creature — Lhurgoyf",
			ManaCost:  "{1}{G}",
			Colors:    []string{"G"},
			Power:     0,
			Toughness: 1,
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Power is the number of card types among cards in all graveyards; toughness is that plus 1.",
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := game.DistinctCardTypesInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
		Text: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
	}
}

func init() {
	tokenTemplates = append(tokenTemplates, printedGapB45TarmogoyfToken)
}

// gapB45DesertWarfareReturnBody is Desert Warfare's delayed body: each
// listed card still in its owner's graveyard is put onto the battlefield
// under the delayed trigger's controller's control.
var gapB45DesertWarfareReturnBody = game.SimpleDelayedBody("desert-warfare/put-listed-onto-the-battlefield", gapB45PutListedOntoBattlefield)

func gapB45PutListedOntoBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// gapB45SatyaSacrificeBody is Satya's delayed trigger: "sacrifice that
// token unless you pay an amount of {E} equal to its mana value". The
// token is the one listed target; the payment is asked as the delayed
// trigger resolves (CR 118.12) and declining, or being short, sacrifices
// it. A token that has gone, or one whose mana value is zero, asks nothing.
var gapB45SatyaSacrificeBody = game.SimpleDelayedBody("satya-aetherflux-genius/sacrifice-unless-energy", gapB45SatyaSacrifice)

func gapB45SatyaSacrifice(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) || c.Controller != item.Controller {
			continue
		}
		mv := c.ManaValue()
		if mv <= 0 {
			continue
		}
		return PayEnergyUnless{
			N:        mv,
			Question: "Satya, Aetherflux Genius — pay " + EnergySymbols(mv) + ", or sacrifice the token?",
			Else:     gapB45SatyaSacrificeToken,
		}.Apply(NewContext(g, item))
	}
	return nil
}

// gapB45SatyaSacrificeToken is the declined half: the token goes. An
// on-disk key.
var gapB45SatyaSacrificeToken = PayUnlessBranches("pay-unless/satya-sacrifice-token", nil,
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.Item.Targets {
			if t.Kind == game.TargetCard && onBattlefield(ctx.Game, t.ID) {
				return SacrificePermanent{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})

// gapB45ColorlessEldraziCopy applies Benthic Anomaly's copy exception to a
// token template: "except its power is equal to the total power of those
// creatures, its toughness is equal to the total toughness of those
// creatures, and it's a colorless Eldrazi creature" (CR 707.9b). The
// exception sets the card type (a creature, only) and the subtype (Eldrazi)
// and keeps the supertypes. A coloured mana cost would colour the token, so
// the template's mana cost is rewritten as the same amount of generic mana:
// the token is colourless, keeps its mana value, and has no coloured symbol.
func gapB45ColorlessEldraziCopy(tok *game.Card, power, toughness int) {
	super, _, _ := game.ParseTypeLine(tok.TypeLine)
	head := append(append([]string(nil), super...), "Creature")
	tok.TypeLine = strings.Join(head, " ") + " — Eldrazi"
	tok.SetCopyExceptionColors() // colourless, and a copied devoid is not copied (CR 707.9d)
	if mv := tok.ManaValue(); mv > 0 {
		tok.ManaCost = "{" + strconv.Itoa(mv) + "}"
	} else {
		tok.ManaCost = ""
	}
	tok.Power, tok.Toughness = power, toughness
	tok.VariableToughness = false
	tok.PrintedPTKnown = true
}
