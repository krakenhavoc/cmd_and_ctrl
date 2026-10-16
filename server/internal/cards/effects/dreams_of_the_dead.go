package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreams of the Dead — Enchantment {3}{U}:
//
//	"{1}{U}: Return target white or black creature card from your
//	 graveyard to the battlefield. That creature gains 'Cumulative upkeep
//	 {2}.' If the creature would leave the battlefield, exile it instead
//	 of putting it anywhere else."
//
// #3869. Three existing pieces in order: the return (ReturnFromGraveyard),
// an indefinite ADR 0093 grant of the CumulativeUpkeepGrant bundle pinned
// to the returned object (CR 611.2a, 611.2c), and Whip of Erebos's
// "exile it instead" replacement (ExileInsteadOfLeavingBattlefield, CR
// 614), which also covers the sacrifice an unpaid upkeep makes. An
// effect that exiles the creature anyway exiles it, and a card that
// comes back later is a new object with neither the grant nor the
// replacement (the 2008-10-01 ruling, CR 400.7).
//
// "white or black" is read through the target walk's word lens (OfColor).
//
// No simplification.
const dreamsOfTheDeadGrant = "dreams-of-the-dead/cumulative-upkeep-2"

func init() {
	Register(Spec{
		OracleID:     "3b6cf935-3b99-4371-a797-6f40b89adeeb",
		Name:         "Dreams of the Dead",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{CumulativeUpkeepGrant(dreamsOfTheDeadGrant, "{2}")},
		Activated: []ActivatedAbility{{
			Label: "{1}{U}: Return target white or black creature card from your graveyard to the battlefield with cumulative upkeep {2}",
			Cost:  ManaCost("{1}{U}"),
			Targets: TargetCardInGraveyard("target white or black creature card in your graveyard",
				Creature(), YouOwn(), Or(OfColor("W"), OfColor("B"))),
			Effect: dreamsOfTheDeadReturn,
		}},
	})
}

func dreamsOfTheDeadReturn(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	id := item.Targets[0].ID
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if !onBattlefield(g, id) {
		return nil
	}
	if err := (GrantAbilitiesFor{
		Target:   id,
		Keys:     []string{dreamsOfTheDeadGrant},
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Dreams of the Dead — the returned creature has cumulative upkeep {2}",
	}).Apply(ctx); err != nil {
		return err
	}
	ExileInsteadOfLeavingBattlefield(g, id, item.Controller,
		"Dreams of the Dead: if it would leave the battlefield, exile it instead")
	return nil
}
