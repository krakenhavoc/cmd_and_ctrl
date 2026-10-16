package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wrenn and Realmbreaker — Legendary Planeswalker — Wrenn {1}{G}{G},
// starting loyalty 3:
//
//	"Lands you control have '{T}: Add one mana of any color.'
//	 +1: Up to one target land you control becomes a 3/3 Elemental
//	 creature with vigilance, hexproof, and haste until your next
//	 turn. It's still a land.
//	 −2: Mill three cards. You may put a permanent card from among the
//	 milled cards into your hand.
//	 −7: You get an emblem with 'You may play lands and cast permanent
//	 spells from your graveyard.'"
//
// The static grant, +1 and −2 all ship complete: the static is the
// ordinary ADR 0093 lands-you-control shape; +1 is one ScopedEffectFor
// registration (ADR 0041 phase 3) with five data Mods — type, subtype,
// base P/T, and the three keywords — pinned to the chosen land for
// DurationUntilYourNextTurn; −2 is MillToZone + mayTakeOneFromAmongThem,
// the Barrowgoyf shape, widened to Permanent() instead of Creature().
//
// The −7 is complete (#3861): the emblem carries two STANDING cast
// permissions over its owner's graveyard (EmblemSpec.CastPermissions,
// #2528, CR 114.3), derived from Player.Emblems on every query exactly
// as Wrenn and Six's retrace emblem is. "Play lands" is the Crucible
// of Worlds permission (LandsOnly: a land play, which spends the
// turn's land drop, CR 305.2a); "cast permanent spells" is a second
// permission over nonland permanent cards (NonLandPermanentOnly),
// cast at their normal timing for their normal cost. Instants and
// sorceries are not permanent cards (CR 110.4) and stay where they
// are. The emblem lasts the rest of the game (CR 114.2).
func init() {
	const grant = "wrenn-and-realmbreaker/any-color"
	Register(Spec{
		OracleID:     "4566fb92-448e-4b3f-9045-9d74323c35d1",
		Name:         "Wrenn and Realmbreaker",
		Completeness: CompletenessFull,
		Emblem: &EmblemSpec{
			Label: "Wrenn and Realmbreaker emblem",
			Text:  "You may play lands and cast permanent spells from your graveyard.",
			CastPermissions: []game.CastPermission{
				{
					Zone:   game.ZoneGraveyard,
					Filter: game.PermissionFilter{LandsOnly: true},
					Label:  "Play a land from your graveyard (Wrenn and Realmbreaker emblem)",
				},
				{
					Zone:   game.ZoneGraveyard,
					Filter: game.PermissionFilter{NonLandPermanentOnly: true},
					Label:  "Cast a permanent spell from your graveyard (Wrenn and Realmbreaker emblem)",
				},
			},
		},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add one mana of any color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			GrantAbilities(landsYouControl, grant),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Up to one target land you control becomes a 3/3 Elemental creature with vigilance, hexproof, and haste until your next turn. It's still a land.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to one target land you control", And(Land(), YouControl())).WithCount(0, 1),
				Effect:  wrennPlusOne,
			},
			{
				Label:  "−2: Mill three cards. You may put a permanent card from among the milled cards into your hand.",
				Cost:   LoyaltyCost(-2),
				Effect: wrennMinusTwo,
			},
			{
				Label: "−7: You get an emblem with \"You may play lands and cast permanent spells from your graveyard.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// wrennPlusOne animates the chosen land as a 3/3 Elemental with
// vigilance, hexproof and haste until the controller's next turn. A
// declined target (nothing chosen, or one that left in response) does
// nothing.
//
// One ScopedEffectFor registration, all data (ADR 0041 phase 3,
// #1497) — no ScopedStatic closure, so the record survives a
// snapshot/restore for as long as the effect lives.
func wrennPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return ScopedEffectFor{
		Target: targets[0].ID,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod("Elemental"),
			game.SetBasePowerMod(3),
			game.SetBaseToughnessMod(3),
			game.AddKeywordsMod("vigilance", "hexproof", "haste"),
		},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Wrenn and Realmbreaker — becomes a 3/3 Elemental until your next turn",
	}.Apply(ctx)
}

// wrennMinusTwo is "Mill three cards. You may put a permanent card
// from among the milled cards into your hand."
func wrennMinusTwo(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 3,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return mayTakeOneFromAmongThem(ctx, milled, Permanent(),
				"Wrenn and Realmbreaker — put a permanent card from among the milled cards into your hand")
		},
	}.Apply(NewContext(g, item))
}
