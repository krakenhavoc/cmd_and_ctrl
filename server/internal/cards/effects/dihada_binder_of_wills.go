package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dihada, Binder of Wills — Legendary Planeswalker — Dihada {1}{R}{W}{B},
// starting loyalty 5 (Dominaria United Commander, #5487):
//
//	"+2: Up to one target legendary creature gains vigilance, lifelink, and
//	 indestructible until your next turn.
//	 −3: Reveal the top four cards of your library. Put any number of legendary
//	 cards from among them into your hand and the rest into your graveyard.
//	 Create a Treasure token for each card put into your graveyard this way.
//	 −11: Gain control of all nonland permanents until end of turn. Untap them.
//	 They gain haste until end of turn.
//	 Dihada, Binder of Wills can be your commander."
//
// The +2's keywords last until the start of its controller's next turn
// (CR 611.2) and are locked to the creature chosen. The −3 reveals all four
// cards (they are public, CR 701.20a), lets the player take any number of the
// legendary ones, then puts the rest into their owners' graveyards through the
// ordinary routed move; a Treasure is made for each card that actually arrived
// in a graveyard (a replacement such as Rest in Peace that exiles one makes
// none). The −11 takes every nonland permanent on the battlefield, untaps
// them all (including ones already under Dihada's controller) and gives each
// haste until end of turn. Commander eligibility is the printed clause
// (internal/deck).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d326de4d-6157-4586-91b8-5668e1381ac3",
		Name:            "Dihada, Binder of Wills",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Up to one target legendary creature gains vigilance, lifelink, and indestructible until your next turn.",
				Cost:    LoyaltyCost(2),
				Targets: TargetCreature("up to one target legendary creature", Legendary()).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.AddKeywordsMod("vigilance", "lifelink", "indestructible")},
						Duration: DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Dihada, Binder of Wills — vigilance, lifelink and indestructible until your next turn",
					}.Apply(ctx)
				},
			},
			{
				Label: "−3: Reveal the top four cards of your library. Put any number of legendary cards from among them into your hand and the rest into your graveyard. Create a Treasure token for each card put into your graveyard this way.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 4,
						"Dihada, Binder of Wills — reveal the top four cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					return TakeFromLibraryToHand{
						Cards:    revealed,
						Match:    Legendary(),
						Max:      0,
						Optional: true,
						Label:    "Dihada, Binder of Wills — put any number of legendary cards from among them into your hand",
						Key:      dihadaTake,
					}.Apply(ctx)
				},
			},
			{
				Label: "−11: Gain control of all nonland permanents until end of turn. Untap them. They gain haste until end of turn.",
				Cost:  LoyaltyCost(-11),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if !c.IsLand() {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (GainControl{Target: id, Label: "Dihada, Binder of Wills — gain control of all nonland permanents"}).Apply(ctx); err != nil {
							return err
						}
					}
					for _, id := range ids {
						if err := (UntapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					for _, id := range ids {
						if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
							Label: "Dihada, Binder of Wills — haste until end of turn"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}

// dihadaTake is the −3's take: the legendary cards taken go to the hand, the
// rest into the graveyard, and a Treasure for each that got there. An on-disk
// key: never renamed, never reused.
var dihadaTake = RegisterTake("choose-cards/dihada-binder-of-wills-take", TakeDef{
	Then: func(g *game.Game, res TakeFromLibraryResult) error {
		err := TakeRestIntoGraveyard(g, res)
		n := 0
		for _, id := range res.Rest {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
				n++
			}
		}
		if n == 0 {
			return err
		}
		ctx := keyedContext(g, res.Player, res.Source)
		if terr := (CreateToken{Controller: res.Player, Template: TreasureToken(), N: n}).Apply(ctx); terr != nil && err == nil {
			err = terr
		}
		return err
	},
})
