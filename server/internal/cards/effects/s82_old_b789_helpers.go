package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b789_helpers.go — Visions, batch 789 (S82 older sets, #5631).

// Old789RustCounter is Corrosion's rust counter (CR 122.1: a counter with no
// rule of its own).
const Old789RustCounter = "rust"

// Old789CorrosionUpkeep is Corrosion's upkeep ability: put a rust counter on
// each artifact the targeted opponent controls, then destroy each artifact
// (whoever controls it) whose mana value is no greater than the number of
// rust counters on it. They can't be regenerated. A target that is no longer
// legal does nothing at all (CR 608.2b).
func Old789CorrosionUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := FirstLegalTargetPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsArtifact() && c.Controller == player {
			mine = append(mine, c.InstanceID)
		}
	}
	for _, id := range mine {
		if err := (AddCounter{Target: id, Kind: Old789RustCounter, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	var doomed []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsArtifact() && c.ManaValue() <= c.Counters[Old789RustCounter] {
			doomed = append(doomed, c.InstanceID)
		}
	}
	for _, id := range doomed {
		if err := (DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// Old789RemoveAllRustCounters is "remove all rust counters from all
// permanents".
func Old789RemoveAllRustCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	type rusted struct {
		id uuid.UUID
		n  int
	}
	var all []rusted
	for _, c := range g.BattlefieldCardsForEffect() {
		if n := c.Counters[Old789RustCounter]; n > 0 {
			all = append(all, rusted{c.InstanceID, n})
		}
	}
	for _, r := range all {
		if err := (AddCounter{Target: r.id, Kind: Old789RustCounter, N: -r.n}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// Old789SongOfBloodBody gives the creature that attacked +N/+0 until end of
// turn, N being the creature cards the mill put into its controller's
// graveyard (the params' Amount, fixed when Song of Blood resolved). An
// on-disk key: never renamed, never reused.
var Old789SongOfBloodBody = game.DelayedBody("s82-old-b789/song-of-blood-attacker-pump",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, id := range ctx.PayloadCards() {
			if !OnBattlefield(g, id) {
				continue
			}
			if err := (BoostUntilEOT{Target: id, Power: p.Amount, Label: "Song of Blood — +1/+0 for each creature card milled"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// Old789SongOfBloodMilled is Song of Blood's keyed tail: count the creature
// cards that reached the graveyard, then watch for attackers this turn.
var Old789SongOfBloodMilled = CardTail("s82-old-b789/song-of-blood-milled", func(ctx *Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	return DelayedOnEvent{
		Label:   "Song of Blood — the attacking creature gets +1/+0 for each creature card milled",
		On:      []game.EventKind{game.EventAttack},
		Body:    Old789SongOfBloodBody,
		Params:  game.EffectParams{Amount: n},
		Repeats: true,
	}.Apply(ctx)
})

// Old789NonblueCantBlockYourCreaturesUnlessPayLife is Heat Wave's life tax:
// CantBlockYourCreaturesUnlessPayLife with the colour word read through the
// printing permanent's text (CR 612.2) rather than hard-coded.
func Old789NonblueCantBlockYourCreaturesUnlessPayLife(life int, label string) game.BlockTax {
	return game.BlockTax{
		Label: label,
		AppliesTo: func(q game.BlockTaxQuery) bool {
			src := q.Source
			if q.Blocker.HasColor(src.TextColor("U")) {
				return false
			}
			for _, a := range q.Blocking {
				if a.Controller == q.Source.Controller {
					return true
				}
			}
			return false
		},
		LifeCost: func(game.BlockTaxQuery) int { return life },
	}
}
