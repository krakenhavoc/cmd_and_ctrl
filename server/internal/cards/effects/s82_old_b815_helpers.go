package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b815_helpers.go — shared pieces of the S82 OLD batch 815 cards
// (Alliances, activated abilities, #5634).

// OldB815TapTheTargets taps every legal target of the resolving ability, a
// permanent at a time (Floodwater Dam's "Tap X target lands").
func OldB815TapTheTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.TapTargetForEffect(t.ID); err != nil {
			return err
		}
	}
	return nil
}

// OldB815AnyTypeTheUntappedLandCouldProduce is Benthic Explorers' "Add one
// mana of any type that land could produce": the land the cost untapped
// (PaidCost.UntappedOthers, CR 400.7j) offers every type of mana it could
// produce right now (CR 106.7: asked of its abilities, not whether it could be
// activated). A land that could produce nothing, or no land, adds none.
func OldB815AnyTypeTheUntappedLandCouldProduce(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	if len(paid.UntappedOthers) == 0 {
		return game.ManaOutput{}
	}
	land, ok := g.LookupCardForEffect(paid.UntappedOthers[0].ID)
	if !ok {
		return game.ManaOutput{}
	}
	seen := map[string]bool{}
	for _, s := range g.ProducibleManaLocked(land) {
		seen[s] = true
	}
	syms := OrderedManaSymbols(seen)
	if len(syms) == 0 {
		return game.ManaOutput{}
	}
	return game.ManaOutput{Text: PipeString(syms)}
}

// OldB815BlackPerManaValueOfTheSacrificed is Soldevi Adnate's "Add an amount
// of {B} equal to the sacrificed creature's mana value": the sum over what the
// cost sacrificed, each as it last existed on the battlefield (CR 608.2h).
func OldB815BlackPerManaValueOfTheSacrificed(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	n := 0
	for _, ref := range paid.SacrificedObjects {
		if info, ok := g.PermanentForEffect(ref); ok {
			n += info.ManaValue
		}
	}
	if n <= 0 {
		return game.ManaOutput{}
	}
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"B"}, N: n}}
}

// OldB815GrantTrampleToTheBand is "Attacking creatures banded with this
// creature gain trample until end of turn": every other attacker still in the
// source's band as the ability resolves, fixed then (CR 611.2c). The source
// itself is not "banded with" itself.
func OldB815GrantTrampleToTheBand(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	for _, id := range g.BandMatesForEffect(item.SourceCardID) {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Urza's Engine — trample until end of turn"}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// OldB815SwornDefenderSwap is Sworn Defender's body: the source's power
// becomes the toughness of the target creature blocking or blocked by it, minus
// 1, and its toughness becomes 1 plus that creature's power, both until end of
// turn. Both are read as the ability resolves (CR 608.2h), before either
// changes, and set the base values (CR 613.4b): counters and pumps still
// apply on top. A target that has left combat is not read (CR 608.2b).
func OldB815SwornDefenderSwap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok || !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	foe, found := g.LookupCardForEffect(t.ID)
	if !found {
		return nil
	}
	power, toughness := foe.CurrentToughness()-1, 1+foe.CurrentPower()
	return UntilEndOfTurn(ctx, item.SourceCardID, nil,
		"Sworn Defender — power becomes that creature's toughness minus 1 and toughness becomes 1 plus its power",
		game.SetBasePTMods(power, toughness)...)
}

// oldB815CylixPick is the pick of Ashnod's Cylix: the
// chosen card stays on top of the library and the others looked at are exiled.
// An on-disk key.
var oldB815CylixPick = game.RegisterChooseCardsThen("choose-cards/old-b815-ashnods-cylix-pick", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Candidates {
		if slicesContainsID(r.Picked, id) {
			continue
		}
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})

func slicesContainsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, have := range ids {
		if have == id {
			return true
		}
	}
	return false
}

// OldB815AshnodsCylix is the ability's body: the targeted player looks at the
// top three cards of their library, puts one of them back on top and exiles
// the rest. They choose; with fewer than three cards they look at what they
// have, and with one the choice is forced.
func OldB815AshnodsCylix(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	looked := g.LookAtTopOfLibraryForEffect(player, 3)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Ashnod's Cylix — put one of them back on top of your library; the rest are exiled",
		Cards:    looked,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  oldB815CylixPick,
	})
	return nil
}
