package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// BecomesShapeshifterUntilEOT is "this artifact becomes a Shapeshifter artifact
// creature with base power and toughness P/T until end of turn" as an
// ability body (Dire Mimic). An ability that resolves after its source
// became a new object does nothing (CR 400.7).
func BecomesShapeshifterUntilEOT(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if SourceIsNewObject(g, item) {
			return nil
		}
		return BecomeCreatureUntilEOT{
			Subtypes: []string{"Shapeshifter"},
			SetPower: power, SetToughness: toughness,
			Label: label,
		}.Apply(NewContext(g, item))
	}
}

// PartyMatching is the size of the largest party (CR 700.8b) among creatures
// whose party types are given as fits (a row per creature, one flag per type
// of FinB9PartyTypes): a maximum bipartite matching between creatures and
// the four types.
func PartyMatching(fits [][4]bool) int {
	owner := [4]int{-1, -1, -1, -1} // type slot → index into fits
	var try func(c int, seen *[4]bool) bool
	try = func(c int, seen *[4]bool) bool {
		for t := 0; t < 4; t++ {
			if !fits[c][t] || seen[t] {
				continue
			}
			seen[t] = true
			if owner[t] < 0 || try(owner[t], seen) {
				owner[t] = c
				return true
			}
		}
		return false
	}
	n := 0
	for c := range fits {
		var seen [4]bool
		if try(c, &seen) {
			n++
		}
	}
	return n
}

// PartySizeInLayers is FinB9PartySize for a static ability that is being
// applied while the layers are computed (Multiclass Baldric): it reads the
// battlefield as the earlier layers have left it, never asking for a
// recompute from inside one. The caller holds the game lock.
func PartySizeInLayers(g *game.Game, player uuid.UUID) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	var fits [][4]bool
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		var f [4]bool
		any := false
		for j, t := range FinB9PartyTypes {
			if c.HasSubtype(t) {
				f[j], any = true, true
			}
		}
		if any {
			fits = append(fits, f)
		}
	}
	return PartyMatching(fits)
}

// ControlsPartyType is "you control a <party type>" for a static ability in
// the layers: a creature of that subtype controlled by `player`.
func ControlsPartyType(g *game.Game, player uuid.UUID, subtype string) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsCreature() && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// CrewedByExactly reports whether the Vehicle object `ref` was crewed this
// turn by exactly n distinct creatures, counting any that have since left
// the battlefield (they still paid the crew cost, CR 702.122a).
func CrewedByExactly(g *game.Game, ref game.ObjectRef, n int) bool {
	seen := map[uuid.UUID]bool{}
	for _, r := range g.CrewRecordsForEffect(ref, false) {
		seen[r.Creature.ID] = true
	}
	return len(seen) == n
}

// CondCrewedByExactlyTwo is "if it was crewed by exactly two creatures" as
// an intervening "if" (CR 603.4): read for the resolving ability's source
// Vehicle. It declares no footprint (CondCorrupted's reason).
type CondCrewedByExactlyTwo struct{}

func (CondCrewedByExactlyTwo) Holds(ctx *Context) bool {
	ref, ok := ctx.SourceRef()
	return ok && CrewedByExactly(ctx.Game, ref, 2)
}

func (CondCrewedByExactlyTwo) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// WhenBecomesCrewedFirstTimeByExactlyTwo is "Whenever this Vehicle becomes
// crewed for the first time each turn, if it was crewed by exactly two
// creatures, …": the first-time trigger with the intervening "if" checked as
// it triggers and as it resolves.
func WhenBecomesCrewedFirstTimeByExactlyTwo(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	t := WhenBecomesCrewedFirstTimeEachTurn(label, effect)
	base := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return base(ev, source, lki, g) && CrewedByExactly(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}, 2)
	}
	return InterveningIf(CondCrewedByExactlyTwo{}, t)
}

// CreaturesOfChooser is the Candidates of a ChoosePermanents that offers the
// creatures its chooser controls, between min and max of them (Fraying Line).
func CreaturesOfChooser(minPick, maxPick int) func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == of && c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
		return ids, minPick, maxPick
	}
}
