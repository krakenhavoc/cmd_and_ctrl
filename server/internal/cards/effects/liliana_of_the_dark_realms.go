package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana of the Dark Realms — Legendary Planeswalker — Liliana
// {2}{B}{B}, starting loyalty 3:
//
//	"+1: Search your library for a Swamp card, reveal it, put it into
//	     your hand, then shuffle.
//	 −3: Target creature gets +X/+X or -X/-X until end of turn, where X
//	     is the number of Swamps you control.
//	 −6: You get an emblem with "Swamps you control have '{T}: Add
//	     {B}{B}{B}{B}.'""
//
// The +1 is a search for any card with the Swamp subtype, not only a
// basic one, so Dryad Arbor-style and nonbasic Swamps are found. It
// reveals, puts the card into the hand and shuffles; a library with no
// Swamp (or a searcher who declines to find one, CR 701.23b) still
// shuffles.
//
// The −3 chooses its sign as it RESOLVES: the target is chosen when the
// ability is put on the stack (CR 601.2c), but "+X/+X or -X/-X" is a
// choice made on resolution (CR 608.2c). X is read when the choice is
// answered, off the effective type line (Urborg, Tomb of Yawgmoth makes
// every land a Swamp), and the pump is then locked in. A target that is
// no longer legal does nothing and asks nothing (CR 608.2b). The answer
// is a registered continuation, so the table stays a restore point while
// the question is open.
//
// The −6 emblem grants a layer-6 bundle, the Tyvar Kell shape: the
// emblem is the grantor, the Swamps are the recipients, and the grant
// lasts as long as the emblem does — for the rest of the game
// (CR 114.2), whether or not Liliana is still on the battlefield, and
// for Swamps that arrive later. The extra {B}{B}{B}{B} is a second mana
// ability beside the Swamp's own {T}: Add {B}, and each Swamp taps for
// only one of them.
//
// No simplification.
func init() {
	lilianaDarkRealmsPump = OptionPickThen("option-pick/liliana-of-the-dark-realms-pump", lilianaDarkRealmsPumped)
	Register(Spec{
		OracleID:     "92301066-8904-41bc-84bd-4aa286cbab9c",
		Name:         "Liliana of the Dark Realms",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for fixtures.
		StartingLoyalty: 3,
		Grants: []AbilityGrant{
			TapForManaGrant(lilianaDarkRealmsGrant, "{B}{B}{B}{B}", "Add {B}{B}{B}{B}", "{T}: Add {B}{B}{B}{B}."),
		},
		Emblem: &EmblemSpec{
			Label: "Liliana of the Dark Realms emblem",
			Text:  "Swamps you control have \"{T}: Add {B}{B}{B}{B}.\"",
			Static: []game.StaticAbility{
				GrantAbilities(lilianaDarkRealmsSwampsYouControl, lilianaDarkRealmsGrant),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: lilianaDarkRealmsIsSwampCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Liliana of the Dark Realms — search for a Swamp card",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−3: Target creature gets +X/+X or -X/-X until end of turn, where X is the number of Swamps you control.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Effect:  lilianaDarkRealmsAskSign,
			},
			{
				Label: "−6: You get an emblem with \"Swamps you control have '{T}: Add {B}{B}{B}{B}.'\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const lilianaDarkRealmsGrant = "liliana-of-the-dark-realms/tap-for-bbbb"

// lilianaDarkRealmsPump is the registered continuation of the −3's
// sign question. Its key is an on-disk identity: never renamed.
var lilianaDarkRealmsPump game.OptionPickThen

// lilianaDarkRealmsIsSwampCard is "a Swamp card": any card with the
// Swamp subtype in the library, basic or not.
func lilianaDarkRealmsIsSwampCard(c game.Card) bool {
	return c.HasSubtype("Swamp")
}

// lilianaDarkRealmsSwampsYouControl is the emblem grant's recipient set:
// every Swamp (by its effective subtypes) the emblem's owner controls.
func lilianaDarkRealmsSwampsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Swamp")
}

// lilianaDarkRealmsAskSign asks whether the target gets +X/+X or -X/-X.
// A target that is already illegal does nothing (CR 608.2b).
func lilianaDarkRealmsAskSign(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	return PickOption{
		Question: "Liliana of the Dark Realms — +X/+X or -X/-X, where X is the number of Swamps you control?",
		Options: []game.ChoiceOption{
			{Label: "+X/+X"},
			{Label: "-X/-X"},
		},
		ThenKey: lilianaDarkRealmsPump,
		Carry:   []uuid.UUID{target},
	}.Apply(ctx)
}

// lilianaDarkRealmsPumped applies the chosen sign with X read now.
func lilianaDarkRealmsPumped(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 || !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	x := swampsControlledBy(ctx.Game, ctx.Controller())
	sign, mark := 1, "+"
	if r.Option.Label == "-X/-X" {
		sign, mark = -1, "-"
	}
	n := strconv.Itoa(x)
	return BoostUntilEOT{
		Target:    r.Carry[0],
		Power:     sign * x,
		Toughness: sign * x,
		Label:     "Liliana of the Dark Realms — " + mark + n + "/" + mark + n,
	}.Apply(ctx)
}
