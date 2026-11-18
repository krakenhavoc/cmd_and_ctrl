package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rekindling Phoenix — Creature — Phoenix {2}{R}{R}, 4/3:
//
//	"Flying
//	 When this creature dies, create a 0/1 red Elemental creature token
//	 with "At the beginning of your upkeep, sacrifice this token and
//	 return target card named Rekindling Phoenix from your graveyard to
//	 the battlefield. It gains haste until end of turn.""
//
// The Phoenix's trigger creates the Elemental, whose own upkeep trigger
// lives on the token (ADR 0083, the Pest's shape) and is the token
// controller's. The target, a card named Rekindling Phoenix in the token
// controller's graveyard, is chosen as the trigger goes on the stack, and
// with none there the trigger is removed. As it resolves the token is
// sacrificed (only the same permanent, still on the battlefield) and the
// card, if it is still in the graveyard, returns to the battlefield, then
// gains haste until end of turn; the return does not depend on the
// sacrifice (the ability is independent of its source once on the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f34be4b8-126d-4b5c-8ebb-23b5e5370012",
		Name:            "Rekindling Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Rekindling Phoenix — create a 0/1 red Elemental creature token with an upkeep ability that brings this card back",
				Do(CreateToken{Template: RekindlingElementalToken(), N: 1})),
		},
	})
}

// RekindlingElementalToken is Rekindling Phoenix's 0/1 red Elemental with
// "At the beginning of your upkeep, sacrifice this token and return target
// card named Rekindling Phoenix from your graveyard to the battlefield. It
// gains haste until end of turn."
func RekindlingElementalToken() game.Card { return tokenFromCatalog(printedRekindlingElementalToken) }

func printedRekindlingElementalToken() tokenTemplate {
	return tokenTemplate{
		Slug: "rekindling-elemental",
		Card: game.Card{
			Name:      "Elemental",
			TypeLine:  "Token Creature — Elemental",
			Power:     0,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Elemental — sacrifice this token and return target card named Rekindling Phoenix from your graveyard to the battlefield; it gains haste until end of turn",
				rprB23RekindleThePhoenix),
				TargetCardInGraveyard("target card named Rekindling Phoenix from your graveyard", YouOwn(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Name == "Rekindling Phoenix" })),
		},
		Text: "At the beginning of your upkeep, sacrifice this token and return target card named Rekindling Phoenix from your graveyard to the battlefield. It gains haste until end of turn.",
	}
}

// rprB23RekindleThePhoenix sacrifices the token (when it is still the same
// permanent on the battlefield), then returns the targeted card and, once
// it is on the battlefield, gives it haste.
func rprB23RekindleThePhoenix(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if onBattlefield(g, item.SourceCardID) && !g.AbilitySourceIsNewObjectForEffect(item) {
		if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, item.Controller, false, ctx.ItemTail(rprB23PhoenixHaste))
}

// rprB23PhoenixHaste is "it gains haste until end of turn". An on-disk key.
var rprB23PhoenixHaste = EnteredTail("rekindling-phoenix/haste", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !onBattlefield(ctx.Game, entered) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: entered, Keywords: []string{"haste"}, Label: "Rekindling Phoenix — haste until end of turn"}.Apply(ctx)
})
