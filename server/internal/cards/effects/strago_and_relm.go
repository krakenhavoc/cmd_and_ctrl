package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strago and Relm — Legendary Creature — Human Wizard {2}{R}, 1/3 (deck
// request #3655):
//
//	"Sketch and Lore — {2}{R}, {T}: Target opponent exiles cards from the
//	 top of their library until they exile an instant, sorcery, or
//	 creature card. You may cast that card without paying its mana cost.
//	 If you cast a creature spell this way, it gains haste and 'At the
//	 beginning of the end step, sacrifice this creature.' Activate only
//	 as a sorcery."
//
// The run is the exile-until primitive (MillToZone with an Until), and
// the card it stops on gets Fevered Suspicion's free-cast permission:
// "{0}" (CR 107.3b locks X at 0), flash timing because the cast happens
// during the resolution (CR 608.2g), and the cascade window that closes
// when you pass priority, leaving the card in exile. Cards exiled on the
// way stay in exile, as printed.
//
// The rider is the permission's (#3768): GainsKeywords gives the
// permanent haste and GainsAbilities the granted end-step sacrifice
// (strago's grant below), with no duration (CR 611.2a), so a creature an
// opponent later steals keeps both. Only a creature spell becomes a
// permanent here, so an instant or sorcery cast this way gains nothing,
// which is the printed "if you cast a creature spell". The granted
// trigger is an ability of the creature, not a delayed trigger: it
// fires at every end step it sees, starting with this turn's.
//
// No simplification.
const stragoAndRelmSacrificeGrant = "strago-and-relm/end-step-sacrifice"

func init() {
	Register(Spec{
		OracleID:     "316e7c28-2b7a-40ae-961b-909644109fe4",
		Name:         "Strago and Relm",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: stragoAndRelmSacrificeGrant,
			Triggered: []game.TriggeredAbility{
				AtEachStep(game.StepEnd, "Strago and Relm — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
					return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
			},
			Text: "At the beginning of the end step, sacrifice this creature.",
		}},
		Activated: []ActivatedAbility{{
			Label:        "Sketch and Lore — {2}{R}, {T}: Target opponent exiles cards from the top of their library until they exile an instant, sorcery, or creature card. You may cast that card without paying its mana cost. If you cast a creature spell this way, it gains haste and \"At the beginning of the end step, sacrifice this creature.\" Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{R}"), TapCost()),
			Targets:      TargetPlayer("target opponent", Opponent()),
			SorcerySpeed: true,
			Effect:       stragoAndRelmSketch,
		}},
	})
}

// stragoAndRelmSketch runs the target opponent's exile-until and offers
// the card it stopped on.
func stragoAndRelmSketch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	return MillToZone{
		Player:   victim,
		To:       game.ZoneExile,
		Until:    game.MillUntil{Ref: stragoAndRelmUntil},
		ThenTail: ctx.ItemTail(stragoAndRelmMilled),
	}.Apply(ctx)
}

// stragoAndRelmUntil is the run's stop clause (#4253 slice 9): the last
// card exiled is a hit.
var stragoAndRelmUntil = game.RegisterMillUntil("until/strago-and-relm-hit", func(landed []game.Card, _ int) bool {
	return len(landed) > 0 && stragoAndRelmHit(landed[len(landed)-1])
})

// stragoAndRelmHit is "an instant, sorcery, or creature card".
func stragoAndRelmHit(c game.Card) bool {
	return c.IsInstant() || c.IsSorcery() || c.IsCreature()
}

// stragoAndRelmMilled is the clause after the mill (#4253 slice 9).
var stragoAndRelmMilled = CardTail("strago-and-relm/after-mill", func(ctx *Context, o game.TailOutcome) error {
	exiled := o.IDs

	if len(exiled) == 0 {
		return nil
	}
	hit := exiled[len(exiled)-1]
	c, ok := ctx.Game.LookupCardForEffect(hit)
	if z := ctx.Game.FindCardZoneForEffect(hit); !ok || z == nil || z.Kind != game.ZoneExile || !stragoAndRelmHit(c) {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:         ctx.Controller(),
		Zone:           game.ZoneExile,
		Cost:           "{0}",
		Timing:         game.TimingFlash,
		CastOnly:       true,
		Duration:       ctx.Game.UntilEndOfTurnDuration(),
		LapseOnPass:    game.LapseStaysInExile,
		GainsKeywords:  []string{"haste"},
		GainsAbilities: []string{stragoAndRelmSacrificeGrant},
		Source:         ctx.Source(),
		SourceName:     "Strago and Relm",
		Label:          "Strago and Relm — cast it without paying its mana cost",
	}, []game.Card{c})
	// #4250: cast during the resolution (CR 608.2g).
	offerResolutionCastOver(ctx.Game, ctx.Controller(), ctx.Source(), "Strago and Relm",
		game.MayCastKeywordFreeCast, declineLeaveInExile, hit)
	return nil
})
