package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "93953926-a644-49bb-9b5a-4c8f19114c7e",
		Name:         "Demolition Field",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Its controller may search for a basic land; you may search for a basic land.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target nonbasic land an opponent controls", Land(), b03Nonbasic(), OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				target := item.Targets[0].ID
				victim, ok := controllerOfTarget(ctx, target)
				if !ok {
					return nil
				}
				if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				if err := (SearchLibrary{
					Player:    victim,
					Predicate: IsBasicLand,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Demolition Field — you may search for a basic land",
				}).Apply(ctx); err != nil {
					return err
				}
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Demolition Field — you may search for a basic land",
				}.Apply(ctx)
			},
		}},
	})
}
