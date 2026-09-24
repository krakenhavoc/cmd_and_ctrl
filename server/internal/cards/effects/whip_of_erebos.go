package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whip of Erebos — Legendary Enchantment Artifact {2}{B}{B} (EDHREC
// rank 710):
//
//	"Creatures you control have lifelink.
//	 {2}{B}{B}, {T}: Return target creature card from your graveyard
//	 to the battlefield. It gains haste. Exile it at the beginning of
//	 the next end step. If it would leave the battlefield, exile it
//	 instead of putting it anywhere else. Activate only as a sorcery."
//
// Every clause is machinery that exists:
//
//   - Lifelink is a Layer 6 grant to creatures you control.
//   - The ability is a sorcery-speed CR 602 ability (SorcerySpeed)
//     with a mana and a tap component and a graveyard target clause.
//   - Haste is GrantKeywordUntilEOT pinned to the returned instance.
//   - "Exile it at the beginning of the next end step" is a delayed
//     trigger carrying the card as its payload (the Waterbender's
//     Restoration shape).
//   - "If it would leave the battlefield, exile it instead" is a
//     turn-scoped replacement (Fog's registry) that redirects any
//     battlefield exit of that one instance to exile — so sacrificing
//     the whipped creature in response to the end-step trigger does
//     NOT put it back in the graveyard for a second whip. Without this
//     clause the card would be stronger than printed; with it, the
//     loop is closed as printed.
//
// One sandbox gap remains, declared below; a second was closed by
// #539 (CR 903.9's exit-primitive rework):
//
//   - CLOSED — a BOUNCE used to bypass the redirect.
//     BounceToHandForEffect now moves the card through
//     routeCardToZoneLocked / executeZoneRouteLocked, the shared exit
//     primitive #539 gave every mover (destroy, sacrifice, bounce,
//     tuck, mill, counter, fizzle), so it runs the same CR 614 window
//     destroy and sacrifice always did. A whipped creature bounced to
//     hand in response is redirected to exile, same as any other
//     exit. See TestB06WhipOfErebosExilesInsteadOfBouncing.
//   - The redirect lives in the turn-scoped registry, which is swept
//     at cleanup. The end-step exile fires before cleanup, so the
//     clause covers the whole printed window; it lapses only if the
//     creature is somehow still on the battlefield after cleanup —
//     which now DOES have a catalog path: Stifle can counter the
//     scheduled "exile the returned creature" delayed trigger like any
//     other triggered ability on the stack (#1211/#1223), leaving the
//     creature to survive past cleanup and go to the graveyard
//     normally instead of exile. Rare, and the one way this ships
//     weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "53987a39-c18c-4c13-b1ea-fd1b2a369f9e",
		Name:         "Whip of Erebos",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If a spell or ability counters the delayed 'exile it' trigger (Stifle, for example) before it resolves, the returned creature can survive past end of turn and later go to the graveyard instead of being exiled.",
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !eotHasAbility(c.Abilities, "lifelink") {
					c.Abilities = append(c.Abilities, "lifelink")
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}{B}, {T}: Return target creature card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step. If it would leave the battlefield, exile it instead. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{B}{B}"), TapCost()),
			Targets:      targetCreatureInYourGraveyard(),
			SorcerySpeed: true,
			Effect:       b06WhipReanimate,
		}},
	})
}

// b06WhipReanimate is the ability body: return, haste, schedule the
// exile, and pin the leaves-the-battlefield redirect to the instance.
func b06WhipReanimate(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	id := item.Targets[0].ID
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Whip of Erebos — the returned creature has haste",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScheduleDelayedTrigger{
		Label: "Whip of Erebos — exile the returned creature",
		Cards: []uuid.UUID{id},
		Body:  exileListedCardsBody,
	}).Apply(ctx); err != nil {
		return err
	}
	// #1221: shared with unearth, which prints the same clause. See
	// exile_instead_of_leaving.go — the caveats below are the
	// helper's, and every card built on it inherits them.
	ExileInsteadOfLeavingBattlefield(g, id, item.Controller,
		"Whip of Erebos: if it would leave the battlefield, exile it instead")
	return nil
}
