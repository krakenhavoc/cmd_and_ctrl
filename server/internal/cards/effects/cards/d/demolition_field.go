package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demolition Field — Land (EDHREC rank 397):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Destroy target nonbasic land an
//	 opponent controls. That land's controller may search their
//	 library for a basic land card, put it onto the battlefield,
//	 then shuffle. You may search your library for a basic land card,
//	 put it onto the battlefield, then shuffle."
//
// Ghost Quarter that replaces itself — the Commander answer to
// Cabal Coffers and Gaea's Cradle that costs no card. Three cost
// components (Myriad Landscape's shape) and a targeted activated
// ability whose clause is three predicates deep: land, nonbasic, an
// opponent's.
//
// Both searches are real "may"s: each is an Optional SearchLibrary,
// which forces the search-chooser prompt so the searcher can decline
// (the S22 chooser; Assassin's Trophy runs the victim's search the
// same way). The victim's prompt is queued first, as printed, and
// both prompts can sit open at once — each is addressed to its own
// player. The basics enter UNTAPPED: the printed text says so, unlike
// Path to Exile's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93953926-a644-49bb-9b5a-4c8f19114c7e",
		Name:         "Demolition Field",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Its controller may search for a basic land; you may search for a basic land.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target nonbasic land an opponent controls", effects.Land(), effects.B03Nonbasic(), effects.OpponentControls()),
			Effect: effects.Do(
				// Its controller's search is open before yours (CR 608.2c, #4155). The
				// engine has checked the target is legal by the time this runs, so the
				// second step needs only the target's kind: the land is gone by then.
				effects.Instruction(func(ctx *effects.Context) error {
					if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					target := ctx.Item.Targets[0].ID
					victim, ok := effects.ControllerOfTarget(ctx, target)
					if !ok {
						return nil
					}
					if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return effects.SearchLibrary{
						Player:    victim,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Demolition Field — you may search for a basic land",
					}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Demolition Field — you may search for a basic land",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
