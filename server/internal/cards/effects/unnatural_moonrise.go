package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unnatural Moonrise — Sorcery {R}{G} (#2586, ADR 0132):
//
//	"It becomes night. Until end of turn, target creature gets +1/+0 and
//	 gains trample and "Whenever this creature deals combat damage to a
//	 player, draw a card."
//	 Flashback {2}{R}{G}"
//
// The creature is the spell's only target, so one that is gone leaves
// the spell without a legal target and it does not resolve (CR 608.2b):
// the night does not come. Otherwise the creature gets the boost, trample and a catalog bundle
// carrying the draw trigger as one duration grant (Herd Heirloom's
// shape, ADR 0093): the trigger is the CREATURE's, so its controller
// draws, and it is gone at the cleanup step. Flashback is the shared
// constructor, so the spell is exiled instead of going to the graveyard
// a second time.
//
// No simplification.
const unnaturalMoonriseDraw = "unnatural-moonrise/draw"

func init() {
	Register(Spec{
		OracleID:         "df33b298-9e4b-4522-96e5-c8d8d6af2bb4",
		Name:             "Unnatural Moonrise",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}{G}")},
		Grants: []AbilityGrant{{
			Key: unnaturalMoonriseDraw,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Unnatural Moonrise — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BecomeNight{}).Apply(ctx); err != nil {
				return err
			}
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: 1, Label: "Unnatural Moonrise — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{unnaturalMoonriseDraw},
				Also:   []game.Mod{game.AddKeywordsMod("trample")},
				Label:  "Unnatural Moonrise — trample and draw on combat damage",
			}.Apply(ctx)
		},
	})
}
