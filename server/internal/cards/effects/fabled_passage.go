package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fabled Passage — Land (EDHREC rank 46):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	card, put it onto the battlefield tapped, then shuffle. Then if
//	you control four or more lands, untap that land."
//
// Evolving Wilds that stops costing you a turn once the game is
// going, which is why it is played over Wilds and Terramorphic
// Expanse in almost every deck that can spare the slot.
//
// The activation cost is Evolving Wilds' — tap and sacrifice, no
// life — so it reuses nothing from fetchlandCost(), which pays a
// life. The land count is taken AFTER the fetch, so the newly
// arrived land counts toward the four; that is the printed reading
// ("Then if you control four or more lands") and it is why a Passage
// cracked with three lands already out comes in untapped.
//
// SANDBOX SIMPLIFICATION — SearchLibrary takes the FIRST basic in
// library order, the same deterministic pick every other search card
// in the catalog makes. That is why the effect peeks at the library
// before searching rather than reading the search's result: the
// primitive returns no handle on what it moved, so "that land" is
// resolved by running the same first-match rule the search will run.
// The two are the same walk over the same slice under the same lock,
// so they cannot disagree.
func init() {
	Register(Spec{
		OracleID: "0c85b8f7-0bd0-4680-9ec5-d4b110460a54",
		Name:     "Fabled Passage",
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fabledPassageFetch,
		}},
	})
}

func fabledPassageFetch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	fetched := firstBasicInLibrary(g, item.Controller)
	if err := (SearchLibrary{
		Player:        item.Controller,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
	}).Apply(ctx); err != nil {
		return err
	}
	// A whiffed search still cost the land; there is nothing to untap.
	if fetched == uuid.Nil {
		return nil
	}
	lands := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Controller == item.Controller {
			lands++
		}
	}
	if lands < 4 {
		return nil
	}
	return UntapTarget{Target: fetched}.Apply(ctx)
}

// firstBasicInLibrary returns the InstanceID SearchLibrary will pick
// for an IsBasicLand predicate — the first match walking the library
// slice from the bottom, which is the order
// SearchLibraryForEffectWithOptions itself walks. uuid.Nil when the
// library holds no basic.
func firstBasicInLibrary(g *game.Game, playerID uuid.UUID) uuid.UUID {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Library == nil {
		return uuid.Nil
	}
	for _, c := range p.Library.Cards {
		if IsBasicLand(c) {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
